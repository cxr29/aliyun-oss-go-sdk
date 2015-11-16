// Copyright 2015 Chen Xianren. All rights reserved.

// Aliyun OSS Go SDK Examples - Multipart Upload File
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"git.oschina.net/cxr29/aliyun-oss-go-sdk"
)

func main() {
	object := oss.NewService("YourAccessKeyId", "YourAccessKeySecret").
		NewBucket("YourBucketName").
		NewObject("YourObjectName")

	// open for reading
	file, err := os.Open("YourFileName")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	// initialize a Multipart Upload
	imu, err := object.InitiateMultipartUpload()
	if err != nil {
		log.Fatalln(err)
	}

	cmu := oss.CompleteMultipartUpload{}

	data := make([]byte, 1<<20) // 1M
	for i := 1; ; i++ {
		n, err := file.Read(data[:])
		if err == nil || err == io.EOF {
			var e error
			if n > 0 {
				cmup := oss.CompleteMultipartUploadPart{PartNumber: i}
				cmup.ETag, e = object.UploadPart(i, imu.UploadId, data[:n])
				if e == nil {
					cmu.Part = append(cmu.Part, cmup)
				} // or retry
			}
			if err == io.EOF && e == nil {
				break
			} else {
				err = e
			}
		}
		if err != nil { // abort the multipart upload
			log.Println(err)
			err = object.AbortMultipartUpload(imu.UploadId)
			if err != nil {
				log.Fatalln(err)
			}
			return
		}
	}

	// complete the multipart upload
	cmur, err := object.CompleteMultipartUpload(imu.UploadId, cmu)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Multipart Upload Etag:", cmur.ETag)
}
