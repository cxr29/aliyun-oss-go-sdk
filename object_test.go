// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const HelloWorld = `// Copyright 2015 Chen Xianren. All rights reserved.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello World!")
}`

func TestObject(t *testing.T) {
	o := newObject()

	_, err := o.Put(HelloWorld)
	if err != errDataTypeNotSupported {
		t.Fatal("expected", errDataTypeNotSupported.Error())
	}

	_, err = o.Put([]byte(HelloWorld))
	e, ok := err.(Error)
	if !(ok && e.Code == "NoSuchBucket") {
		t.Fatal("expected NoSuchBucket")
	}

	fatal(t, o.Bucket.Put())

	etag, err := o.Put(strings.NewReader(HelloWorld))
	fatal(t, err)
	if etag == "" {
		t.Fatal("expected ETag")
	}

	var v []byte
	fatal(t, o.Get(&v))

	if HelloWorld != string(v) {
		t.Fatal("expected HelloWorld")
	}

	fatal(t, o.Delete())

	fatal(t, o.Bucket.Delete())
}

func TestObjectFile(t *testing.T) {
	o := newObject()

	f, err := ioutil.TempFile("", "aliyun-oss-go-sdk-")
	fatal(t, err)
	defer f.Close()
	defer os.Remove(f.Name())

	_, err = io.Copy(f, strings.NewReader(HelloWorld))
	fatal(t, err)

	_, err = f.Seek(0, os.SEEK_SET)
	fatal(t, err)

	fatal(t, o.Bucket.Put())

	etag, err := o.Put(f)
	fatal(t, err)
	if etag == "" {
		t.Fatal("expected ETag")
	}

	_, err = f.Seek(0, os.SEEK_SET)
	fatal(t, err)

	fatal(t, o.Get(f))

	_, err = f.Seek(0, os.SEEK_SET)
	fatal(t, err)

	v, err := ioutil.ReadAll(f)
	fatal(t, err)

	if HelloWorld != string(v) {
		t.Fatal("expected HelloWorld")
	}

	fatal(t, o.Delete())

	fatal(t, o.Bucket.Delete())
}

func TestObjectACL(t *testing.T) {
	o := newObject()

	fatal(t, o.Bucket.Put())
	etag, err := o.Put(bytes.NewBuffer([]byte(HelloWorld)))
	fatal(t, err)
	if etag == "" {
		t.Fatal("expected ETag")
	}

	o.ACL, err = o.GetACL()
	fatal(t, err)
	equal(t, "acl", "default", o.ACL)

	o.ACL = "p"
	err = o.PutACL()
	e, ok := err.(Error)
	if !(ok && e.Code == "InvalidArgument") {
		t.Fatal("expected InvalidArgument")
	}

	o.ACL = ACLPublicRead
	fatal(t, o.PutACL())
	o.ACL = "r"
	o.ACL, err = o.GetACL()
	fatal(t, err)
	equal(t, "acl", ACLPublicRead, o.ACL)

	o.ACL = ACLPublicReadWrite
	fatal(t, o.PutACL())
	o.ACL = "rw"
	o.ACL, err = o.GetACL()
	fatal(t, err)
	equal(t, "acl", ACLPublicReadWrite, o.ACL)

	fatal(t, o.Delete())
	fatal(t, o.Bucket.Delete())
}

func TestObjectCopy(t *testing.T) {
	src := newObject()
	dst := Object{
		Bucket: src.Bucket,
		Name:   "copy-" + src.Name,
	}

	fatal(t, src.Bucket.Put())

	etag, err := src.Put([]byte(HelloWorld))
	fatal(t, err)
	if etag == "" {
		t.Fatal("expected ETag")
	}

	cor, err := dst.Copy(src)
	equal(t, "ETag", etag, cor.ETag)

	var v []byte
	fatal(t, dst.Get(&v))

	if HelloWorld != string(v) {
		t.Fatal("expected HelloWorld")
	}

	fatal(t, src.Delete())
	fatal(t, dst.Delete())

	fatal(t, src.Bucket.Delete())
}

func TestObjectAppend(t *testing.T) {
	o := newObject()

	fatal(t, o.Bucket.Put())

	n := int64(len(HelloWorld))
	h := n / 2

	next, crc, etag, err := o.Append(0, []byte(HelloWorld[:h]))
	fatal(t, err)
	equal(t, "next append position", h, next)
	if crc == "" {
		t.Fatal("expected CRC64")
	}
	if etag == "" {
		t.Fatal("expected ETag")
	}

	next, crc, etag, err = o.Append(next, []byte(HelloWorld[h:]))
	fatal(t, err)
	equal(t, "next append position", n, next)
	if crc == "" {
		t.Fatal("expected CRC64")
	}
	if etag == "" {
		t.Fatal("expected ETag")
	}

	var v []byte
	fatal(t, o.Get(&v))

	if HelloWorld != string(v) {
		t.Fatal("expected HelloWorld")
	}

	fatal(t, o.Delete())

	fatal(t, o.Bucket.Delete())
}

func TestObjectHead(t *testing.T) {
	o := newObject()

	fatal(t, o.Bucket.Put())

	header := Params{}
	header.Set("x-oss-meta-hello", "world")

	etag, err := o.Put(strings.NewReader(HelloWorld), header)
	fatal(t, err)
	if etag == "" {
		t.Fatal("expected ETag")
	}

	h, err := o.Head()
	fatal(t, err)
	equal(t, "head meta", h.Get("x-oss-meta-hello"), "world")

	var v []byte
	fatal(t, o.Get(&v))

	if HelloWorld != string(v) {
		t.Fatal("expected HelloWorld")
	}

	fatal(t, o.Delete())

	fatal(t, o.Bucket.Delete())
}

func TestCORS(t *testing.T) {
	o := newObject()

	fatal(t, o.Bucket.Put())

	etag, err := o.Put(bytes.NewReader([]byte(HelloWorld)))
	fatal(t, err)
	if etag == "" {
		t.Fatal("expected ETag")
	}

	v := new(CORSConfiguration)
	r := CORSRule{
		AllowedOrigin: "*",
		AllowedMethod: "GET",
		AllowedHeader: "*",
		ExposeHeader:  "x-oss-test",
		MaxAgeSeconds: 10,
	}
	v.CORSRule = append(v.CORSRule, r)
	fatal(t, o.Bucket.PutCORS(*v))
	v = nil

	v, err = o.Bucket.GetCORS()
	fatal(t, err)
	equal(t, "CORSRule", 1, len(v.CORSRule))
	equal(t, "CORSRule.AllowedOrigin", r.AllowedOrigin, v.CORSRule[0].AllowedOrigin)
	equal(t, "CORSRule.AllowedMethod", r.AllowedMethod, v.CORSRule[0].AllowedMethod)
	equal(t, "CORSRule.AllowedHeader", r.AllowedHeader, v.CORSRule[0].AllowedHeader)
	equal(t, "CORSRule.ExposeHeader", r.ExposeHeader, v.CORSRule[0].ExposeHeader)
	equal(t, "CORSRule.MaxAgeSeconds", r.MaxAgeSeconds, v.CORSRule[0].MaxAgeSeconds)

	header := Params{}
	header.Set("Origin", "http://www.cxr29.com")
	header.Set("Access-Control-Request-Method", "GET")

	h, err := o.Options(header)
	fatal(t, err)
	equal(t, "Access-Control-Allow-Origin", h.Get("Access-Control-Allow-Origin"), r.AllowedOrigin)
	equal(t, "Access-Control-Allow-Methods", h.Get("Access-Control-Allow-Methods"), r.AllowedMethod)
	equal(t, "Access-Control-Expose-Headers", h.Get("Access-Control-Expose-Headers"), r.ExposeHeader)

	fatal(t, o.Bucket.DeleteCORS())

	v, err = o.Bucket.GetCORS()
	e, ok := err.(Error)
	if !(ok && e.Code == "NoSuchCORSConfiguration") {
		t.Fatal("expected NoSuchCORSConfiguration")
	}

	_, err = o.Options(header)
	e, ok = err.(Error)
	if !(ok && e.Code == "AccessForbidden") {
		t.Fatal("expected AccessForbidden")
	}

	fatal(t, o.Delete())

	fatal(t, o.Bucket.Delete())
}

func TestMultipartUpload(t *testing.T) {
	o := newObject()

	fatal(t, o.Bucket.Put())

	const n = 100*1024 + 2
	part := make([]byte, n)
	copy(part, HelloWorld)

	imu, err := o.InitiateMultipartUpload()
	fatal(t, err)
	if imu.UploadId == "" {
		t.Fatal("expected UploadId")
	}

	cmu := CompleteMultipartUpload{}

	etag, err := o.UploadPart(1, imu.UploadId, part[:n-1])
	fatal(t, err)
	if etag == "" {
		t.Fatal("expected ETag")
	}
	cmu.Part = append(cmu.Part, CompleteMultipartUploadPart{1, etag})

	etag, err = o.UploadPart(2, imu.UploadId, part[n-1:])
	fatal(t, err)
	if etag == "" {
		t.Fatal("expected ETag")
	}
	cmu.Part = append(cmu.Part, CompleteMultipartUploadPart{2, etag})

	cmur, err := o.CompleteMultipartUpload(imu.UploadId, cmu)
	fatal(t, err)
	if cmur.ETag == "" {
		t.Fatal("expected ETag")
	}

	equal(t, "Bucket", o.Bucket.Name, cmur.Bucket)
	equal(t, "Key", o.Name, cmur.Key)

	var v []byte
	fatal(t, o.Get(&v))

	if !bytes.Equal(part, v) {
		t.Fatal("expected Equal")
	}

	fatal(t, o.Delete())

	imu, err = o.InitiateMultipartUpload()
	fatal(t, err)
	if imu.UploadId == "" {
		t.Fatal("expected UploadId")
	}

	src := Object{
		Bucket: o.Bucket,
		Name:   "source-" + o.Name,
	}

	etag, err = src.Put(part)
	fatal(t, err)
	if etag == "" {
		t.Fatal("expected ETag")
	}

	cpr, err := o.UploadPartCopy(1, imu.UploadId, src)
	fatal(t, err)
	equal(t, "ETag", etag, cpr.ETag)

	lmur, err := o.Bucket.ListMultipartUploads()
	fatal(t, err)
	found := false
	for _, i := range lmur.Upload {
		if i.UploadId == imu.UploadId && i.Key == o.Name {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("ListMultipartUploads not found UploadId")
	}

	lpr, err := o.ListParts(imu.UploadId)
	fatal(t, err)
	equal(t, "Part", 1, len(lpr.Part))
	equal(t, "Part.Number", 1, lpr.Part[0].PartNumber)
	equal(t, "Part.ETag", etag, lpr.Part[0].ETag)
	equal(t, "Part.Size", len(part), lpr.Part[0].Size)

	err = o.AbortMultipartUpload(imu.UploadId)
	fatal(t, err)

	_, err = o.UploadPart(2, imu.UploadId, []byte(HelloWorld))
	e, ok := err.(Error)
	if !(ok && e.Code == "NoSuchUpload") {
		t.Fatal("expected NoSuchUpload")
	}

	fatal(t, src.Delete())

	fatal(t, o.Bucket.Delete())
}
