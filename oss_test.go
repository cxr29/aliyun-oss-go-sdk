// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"os"
	"runtime"
	"strconv"
	"testing"
)

func init() {
	pause, _ = strconv.Atoi(os.Getenv("OSSTestPause"))
}

func stack(t *testing.T) {
	buf := make([]byte, 10*1024)
	t.Logf("%s", buf[:runtime.Stack(buf, false)])
}

func equal(t *testing.T, what string, expected, got interface{}) {
	if expected != got {
		stack(t)
		t.Fatal(what, "expected", expected, "but got", got)
	}
}

func fatal(t *testing.T, err error) {
	if err != nil {
		stack(t)
		t.Fatal(err)
	}
}

func newService() Service {
	return Service{
		Unsafe:          os.Getenv("OSSTestUnsafe") == "true",
		Domain:          os.Getenv("OSSTestDomain"),
		AccessKeyId:     os.Getenv("OSSTestAccessKeyId"),
		AccessKeySecret: os.Getenv("OSSTestAccessKeySecret"),
		SecurityToken:   os.Getenv("OSSTestSecurityToken"),
	}
}

func newBucket() Bucket {
	return Bucket{
		Service: newService(),
		Name:    os.Getenv("OSSTestBucket"),
	}
}

func newObject() Object {
	o := Object{
		Bucket: newBucket(),
		Name:   "aliyun-oss-go-sdk.go",
	}
	return o
}
