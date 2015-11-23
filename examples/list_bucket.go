// Copyright 2015 Chen Xianren. All rights reserved.

// Aliyun OSS Go SDK Examples - List Bucket
package main

import (
	"fmt"
	"log"

	"github.com/cxr29/aliyun-oss-go-sdk"
)

func main() {
	service := oss.NewService("YourAccessKeyId", "YourAccessKeySecret")

	{ // list all buckets
		buckets, err := service.ListBucket() // max-keys default 100, max 1000.
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("All buckets of owner %s:\n", buckets.Owner.ID)
		for i, bucket := range buckets.Buckets.Bucket {
			fmt.Printf("\tBucket %d, Name: %s, Location: %s, CreationDate: %s\n",
				i+1, bucket.Name, bucket.Location, bucket.CreationDate)
		}
	}

	{ // list go-* buckets one by one
		query := make(oss.Params, 3)
		query.Set("prefix", "go-")
		query.Set("max-keys", "1")

		buckets, err := service.ListBucket(nil, query)
		if err != nil {
			log.Fatalln(err)
		}

		for result := buckets; result.IsTruncated; {
			query.Set("marker", result.NextMarker)
			result, err = service.ListBucket(nil, query)
			if err != nil {
				log.Fatalln(err)
			}
			buckets.Buckets.Bucket = append(buckets.Buckets.Bucket, result.Buckets.Bucket...)
		}

		fmt.Printf("All go-* buckets of owner %s:\n", buckets.Owner.ID)
		for i, bucket := range buckets.Buckets.Bucket {
			fmt.Printf("\tBucket %d, Name: %s, Location: %s, CreationDate: %s\n",
				i+1, bucket.Name, bucket.Location, bucket.CreationDate)
		}
	}
}
