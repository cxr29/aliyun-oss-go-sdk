// Copyright 2015 Chen Xianren. All rights reserved.

// Aliyun OSS Go SDK Examples - Object Range Get
package main

import (
	"log"
	"os"

	"github.com/cxr29/aliyun-oss-go-sdk"
)

func main() {
	object := oss.NewService("YourAccessKeyId", "YourAccessKeySecret").
		NewBucket("YourBucketName").
		NewObject("YourObjectName")

	// open for writing
	file, err := os.OpenFile("YourFileName", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	var first, length int64 = 0, 1 << 20 // 1M
	for {
		var data []byte
		rl, il, err := object.Range(first, length, &data)
		if err != nil { // or retry
			log.Fatalln(err)
		}

		_, err = file.Write(data)
		if err != nil {
			log.Fatalln(err)
		}

		first += rl
		if first == il {
			break
		} else if first > il || length == 0 {
			log.Fatalln("somthing wrong")
		}

		if first+length > il { // last range
			length = 0
		}
	}
}
