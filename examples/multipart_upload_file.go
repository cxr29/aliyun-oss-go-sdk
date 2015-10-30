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

	for i := 1; ; i++ {
		data := make([]byte, 1<<20) // 1M
		var etag string
		n, err := file.Read(data)
		if err == io.EOF {
			break
		} else if err == nil {
			etag, err = object.UploadPart(i, imu.UploadId, data[:n])
		}
		if err != nil { // abort the multipart upload or retry
			log.Println(err)
			err = object.AbortMultipartUpload(imu.UploadId)
			if err != nil {
				log.Fatalln(err)
			}
			return
		}
		cmu.Part = append(cmu.Part, oss.CompleteMultipartUploadPart{i, etag})
	}

	// complete the multipart upload
	cmur, err := object.CompleteMultipartUpload(imu.UploadId, cmu)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Multipart Upload Etag:", cmur.ETag)
}
