// Copyright 2015 Chen Xianren. All rights reserved.

// Aliyun OSS Go SDK Examples - Object Operations
package main

import (
	"bytes"
	"fmt"
	"log"

	"git.oschina.net/cxr29/aliyun-oss-go-sdk"
)

func main() {
	object := oss.NewService("YourAccessKeyId", "YourAccessKeySecret").
		NewBucket("YourBucketName").
		NewObject("YourObjectName")

	// create the object
	var data = []byte("Hello World!")
	etag, err := object.Put(data)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Object ETag:", etag)

	// change the acl
	object.ACL = oss.ACLPublicRead
	err = object.PutACL()
	if err != nil {
		log.Fatalln(err)
	}

	// copy a object
	copyObject := oss.Object{
		Bucket: object.Bucket,
		Name:   "YourAnotherObjectName",
		ACL:    oss.ACLPrivate,
	}
	cor, err := copyObject.Copy(object)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Copy object ETag:", cor.ETag)
	fmt.Println("Copy object LastModified:", cor.LastModified)

	// get object data
	var copyData []byte
	err = copyObject.Get(&copyData)
	if err != nil {
		log.Fatalln(err)
	}
	if !bytes.Equal(data, copyData) {
		log.Fatalln("Copy object corrupt")
	}

	// delete the objects
	err = object.Delete()
	if err != nil {
		log.Fatalln(err)
	}
	err = copyObject.Delete()
	if err != nil {
		log.Fatalln(err)
	}
}
