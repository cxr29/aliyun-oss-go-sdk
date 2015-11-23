Aliyun OSS Go SDK
=================

[![GoDoc](https://godoc.org/github.com/cxr29/aliyun-oss-go-sdk?status.svg)](https://godoc.org/github.com/cxr29/aliyun-oss-go-sdk)
[![Build Status](https://drone.io/github.com/cxr29/aliyun-oss-go-sdk/status.png)](https://drone.io/github.com/cxr29/aliyun-oss-go-sdk/latest)

### Feature
- All OSS API
- Multipart Upload
- Range Get
- Full Tested
- Simple and Easy to use
- ...

### Install
```$ go get github.com/cxr29/aliyun-oss-go-sdk```

### Usage
```
// Service:
s := oss.NewService("YourAccessKeyId", "YourAccessKeySecret")
// or new service use struct literals
//s := oss.Service{
//	AccessKeyId:     "YourAccessKeyId",
//	AccessKeySecret: "YourAccessKeySecret",
//	...
//}

buckets, err := s.ListBucket() // list my bucket

// Bucket:
b := s.NewBucket("YourBucketName")
// or new bucket use struct literals
//b := oss.Bucket{
//	Service: s,
//	Name:    "YourBucketName",
//	...
//}

err = b.Put()                     // create new bucket
b.Location, err = b.GetLocation() // get and record the location
err = b.Delete()                  // delete the bucket

objects, err := b.ListObject() // list my object

// Object:
o := b.NewObject("YourObjectName")
// or new object use struct literals
//o := oss.Object{
//	Bucket: b,
//	Name:   "YourObjectName",
//	...
//}

var v []byte
// put the bytes as the object data
err = o.Put(v)
// get the object data to the bytes
err = o.Get(&v)

var f *os.File
// open the file for reading then put the file as the object data
err = o.Put(f)
// open the file for writing then get the object data to the file
err = o.Get(f)

// get and record the acl
o.ACL, err = o.GetACL()

// put change the acl
o.ACL = oss.ACLPublicRead
err = o.PutACL()

// delete the object
err = o.Delete()

// Multiupload Upload:
// initialize a Multipart Upload
imu, err := o.InitiateMultipartUpload()

cmu := oss.CompleteMultipartUpload{}

// upload a part
etag, err := o.UploadPart(1, imu.UploadId, part)
cmu.Part = append(cmu.Part, oss.CompleteMultipartUploadPart{1, etag})
// or upload a part copy from the source object
cpr, err := o.UploadPartCopy(1, imu.UploadId, source)
cmu.Part = append(cmu.Part, oss.CompleteMultipartUploadPart{1, cpr.ETag})

// ...

// complte the multipart upload
cmur, err := o.CompleteMultipartUpload(imu.UploadId, cmu)
// or abort the multipart upload
err = o.AbortMultipartUpload(imu.UploadId)
```
More see the examples.

### API Doc
https://godoc.org/github.com/cxr29/aliyun-oss-go-sdk

### Test
```
$ export OSSTestAccessKeyId=YourAccessKeyId
$ export OSSTestAccessKeySecret=YourAccessKeySecret
$ export OSSTestBucket=YourBucketName
$ go test -test.v .
```
OSSTestPause pause in seconds when send request to OSS, because run test so fast will failed.
OSSTestUnsafe, OSSTestDomain and OSSTestSecurityToken also supported.

### Author
Chen Xianren &lt;cxr29@foxmail.com&gt;
