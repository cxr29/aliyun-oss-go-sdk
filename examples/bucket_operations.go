// Copyright 2015 Chen Xianren. All rights reserved.

// Aliyun OSS Go SDK Examples - Bucket Operations
package main

import (
	"log"

	"github.com/cxr29/aliyun-oss-go-sdk"
)

func main() {
	bucket := oss.Bucket{
		Service:  oss.NewService("YourAccessKeyId", "YourAccessKeySecret"),
		Name:     "YourBucketName",
		ACL:      oss.ACLPublicRead,
		Location: oss.LocationCNBeijing,
	}

	// create the public read bucket at beijing
	err := bucket.Put()
	if err != nil {
		log.Fatalln(err)
	}

	// get the location
	location, err := bucket.GetLocation()
	if err != nil {
		log.Fatalln(err)
	}
	if location != bucket.Location {
		log.Fatalln("The bucket location is not beijing")
	}

	// change the acl
	bucket.ACL = oss.ACLPrivate
	err = bucket.PutACL()
	if err != nil {
		log.Fatalln(err)
	}

	// open the bucket server access logs
	logging := new(oss.BucketLoggingStatus)
	logging.LoggingEnabled.TargetBucket = bucket.Name // store to itself
	logging.LoggingEnabled.TargetPrefix = "log-"
	err = bucket.PutLogging(*logging)
	if err != nil {
		log.Fatalln(err)
	}

	// get the bucket logging status
	bls, err := bucket.GetLogging()
	if err != nil {
		log.Fatalln(err)
	}
	if !(bls.LoggingEnabled.TargetBucket == logging.LoggingEnabled.TargetBucket &&
		bls.LoggingEnabled.TargetPrefix == logging.LoggingEnabled.TargetPrefix) {
		log.Fatalln("The bucket logging status is incorrect")
	}

	// close the bucket server access logs
	err = bucket.DeleteLogging()
	if err != nil {
		log.Fatalln(err)
	}

	// The Website, Referrer, Lifcycle are same as Logging
	// ...

	// remove the bucket if no objects, otherwise return BucketNotEmpty Error
	err = bucket.Delete()
	if err != nil {
		log.Fatalln(err)
	}
}
