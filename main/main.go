package main

import (
	"ImageScan/handler"
	"log"
)

/*
Go Main package
*/

func main() {
	log.Println("Image Scan Request received..!!")
	handler.Start()
}
