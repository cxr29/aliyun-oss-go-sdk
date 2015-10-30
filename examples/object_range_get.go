// Copyright 2015 Chen Xianren. All rights reserved.

// Aliyun OSS Go SDK Examples - Object Range Get
package main

import (
	"fmt"
	"log"
	"os"

	"git.oschina.net/cxr29/aliyun-oss-go-sdk"
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

	const size = 1 << 20 // 1M

	header := oss.Params{}
	next := 0
	for {
		header.Set("Range", fmt.Sprintf("bytes=%d-%d", next, next+size))

		res, err := object.GetResponse("GET", nil, header)
		if err != nil { // or retry
			log.Fatalln(err)
		}

		cr := res.Header.Get("Content-Range")

		// whole file
		if cr == "" {
			if next == 0 && res.StatusCode == 200 {
				err = oss.ReadBody(res, file)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				log.Fatalln("range get failed")
			}
			break
		}

		var start, end, total int
		_, err = fmt.Sscanf(cr, "bytes %d-%d/%d", start, end, total)
		if !(err == nil && start == next && end <= next+size && start <= end && end < total) {
			log.Fatalln("range get corrupt")
		}

		var data []byte
		err = oss.ReadBody(res, data)
		if err != nil {
			log.Fatalln(err)
		}

		_, err = file.Write(data)
		if err != nil {
			log.Fatalln(err)
		}

		next += end + 1
		if next == total {
			break
		} else if next > total {
			log.Fatalln("no more range")
		}
	}
}
