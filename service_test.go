// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"net/http"
	"testing"
	"time"
)

func TestServiceListBucket(t *testing.T) {
	s := newService()
	_, err := s.ListBucket()
	fatal(t, err)

	s.Unsafe = true
	s.Domain = GetDomain(LocationCNShanghai, false)
	_, err = s.ListBucket()
	fatal(t, err)
}

var (
	ss = Service{
		Domain:          GetDomain(LocationCNHangzhou, false),
		AccessKeyId:     "44CF9590006BF252F707",
		AccessKeySecret: "OtxrzxIsfpFjA7SwPzILwy8Bw21TLhquhboDYROV",
	}
	sb = Bucket{
		Service: ss,
		Name:    "oss-example",
	}
)

func TestServiceSignatureHeader(t *testing.T) {
	o := Object{
		Bucket: sb,
		Name:   "nelson",
	}

	header := Params{}
	header.Set("X-OSS-Meta-Author", "foo@bar.com")
	header.Set("X-OSS-Magic", "abracadabra")
	header.Set("Date", "Thu, 17 Nov 2005 18:49:58 GMT")
	header.Set("Content-Type", "text/html")
	header.Set("Content-Md5", "ODBGOERFMDMzQTczRUY3NUE3NzA5QzdFNUYzMDQxNEM=")

	req, err := o.GetRequest("PUT", nil, header)
	if err != nil {
		t.Fatal(err)
	}

	o.Signature(req, 0)

	if req.Header.Get("Authorization") != "OSS 44CF9590006BF252F707:26NBxoKdsyly4EDv6inkoDft/yA=" {
		t.Fatal("signature header failed")
	}
}

func TestServiceSignatureUrl(t *testing.T) {
	o := Object{
		Bucket: sb,
		Name:   "oss-api.pdf",
	}

	header := Params{}
	header.Set("Date", time.Unix(1141889060, 0).UTC().Format(http.TimeFormat))

	req, err := o.GetRequest("GET", nil, header)
	if err != nil {
		t.Fatal(err)
	}

	o.Signature(req, 60)

	if req.URL.RawQuery != "OSSAccessKeyId=44CF9590006BF252F707&Expires=1141889120&Signature=EwaNTn1erJGkimiJ9WmXgwnANLc%3D" {
		t.Fatal("signature url failed", req.URL.RawQuery)
	}
}
