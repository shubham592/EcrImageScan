package handler

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

/*
Handler to handle generate Image Scan Request
*/

type ScanSpec struct {
	// ID is a unique identifier for the scan spec
	ID string `json:"id"`
	// CreationTime is the UTC timestamp of when the scan spec was created
	CreationTime string `json:"created"`
	// Region specifies the region the repository is in
	Region string `json:"region"`
	// RegistryID specifies the registry ID
	RegistryID string `json:"registry"`
	// Repository specifies the repository name
	Repository string `json:"repository"`
	// Tags to take into consideration, if empty, all tags will be scanned
	Tags []string `json:"tags"`
}

func StartScan(svc *ecr.ECR, scanspec ScanSpec) (*ecr.StartImageScanOutput, error) {

	scaninput := &ecr.StartImageScanInput{
		RepositoryName: &scanspec.Repository,
		RegistryId:     &scanspec.RegistryID,
	}

	var result *ecr.StartImageScanOutput

	switch len(scanspec.Tags) {
	case 0: // empty list of tags, scan all tags:
		fmt.Printf("DEBUG:: scanning all tags for repo %v\n", scanspec.Repository)
		lio, err := svc.ListImages(&ecr.ListImagesInput{
			RepositoryName: &scanspec.Repository,
			RegistryId:     &scanspec.RegistryID,
			Filter: &ecr.ListImagesFilter{
				TagStatus: aws.String("TAGGED"),
			},
		})
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		for _, iid := range lio.ImageIds {
			scaninput.ImageId = iid
			result, err := svc.StartImageScan(scaninput)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			fmt.Printf("DEBUG:: result for tag %v: %v\n", *iid.ImageTag, result)
			//return result, nil
		}

	default: // iterate over the tags specified in the config:
		fmt.Printf("DEBUG:: scanning tags %v for repo %v\n", scanspec.Tags, scanspec.Repository)
		for _, tag := range scanspec.Tags {
			scaninput.ImageId = &ecr.ImageIdentifier{
				ImageTag: aws.String(tag),
			}
			result, err := svc.StartImageScan(scaninput)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			fmt.Printf("DEBUG:: result for tag %v: %v\n", tag, result)
			//return result, nil
		}
	}
	return result, nil
}

func Start() error {

	//Put the image details here which you want to scan
	log.Println("Started Image Scan")
	region := "us-east-2"
	s := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := ecr.New(s)

	listRepositories, er := svc.DescribeRepositories(&ecr.DescribeRepositoriesInput{})
	if er != nil {
		log.Println(er)
	}

	listRepo := listRepositories.Repositories
	for _, iid := range listRepo {

		listImages, eerr := svc.ListImages(&ecr.ListImagesInput{
			RegistryId:     iid.RegistryId,
			RepositoryName: iid.RepositoryName,
		})

		if eerr != nil {
			log.Println(eerr)
		}
		imageList := listImages.ImageIds

		for _, imageid := range imageList {
			scanSpecification := ScanSpec{
				ID:         *imageid.ImageDigest,
				Repository: *iid.RepositoryName,
				Region:     region,
				Tags:       []string{},
			}

			_, err := StartScan(svc, scanSpecification)
			if err != nil {
				fmt.Println(err)
				return err
			}

			ReportOutput,_:= svc.DescribeImageScanFindings(&ecr.DescribeImageScanFindingsInput{
				ImageId:        &ecr.ImageIdentifier{
					ImageDigest: imageid.ImageDigest,
					ImageTag:    imageid.ImageTag,
				},
				RepositoryName: iid.RepositoryName,
			})

			log.Println("Scan Output", ReportOutput.GoString())

		}
	}
	return nil
}
