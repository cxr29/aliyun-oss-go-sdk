// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	errAccessKeyRequired    = errors.New("access key required")
	errBucketNameRequired   = errors.New("bucket name required")
	errBucketNameInvalid    = errors.New("bucket name invalid")
	errObjectNameRequired   = errors.New("object name required")
	errObjectNameInvalid    = errors.New("object name invalid")
	errDataTypeNotSupported = errors.New("data type not supported")
	errUploadIdRequired     = errors.New("upload id required")
	errPartNumberInvalid    = errors.New("part number invalid")
	errSourceObjectInvalid  = errors.New("source object invalid")
)

// Service represents Aliyun Object Storage Service,
// the AccessKeyId and the AccessKeySecret are required.
type Service struct {
	Unsafe          bool
	Domain          string
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string // STS
}

// NewService returns a new Service given a accessKeyId and accessKeySecret.
func NewService(accessKeyId, accessKeySecret string) Service {
	return Service{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
}

// NewBucket returns a new Bucket given a bucketName from the service.
func (s Service) NewBucket(bucketName string) Bucket {
	return Bucket{
		Service: s,
		Name:    bucketName,
	}
}

// Scheme returns http if the Unsafe is ture otherwise returns https.
func (s Service) Scheme() string {
	if s.Unsafe {
		return "http"
	}
	return "https"
}

// Host returns the OSS access domain.
func (s Service) Host() string {
	if s.Domain == "" {
		return GetDomain("", false)
	}
	return s.Domain
}

// Do sends an HTTP request to OSS and read the HTTP response to v.
//
// The first optional Params is for Header, the second is for Query.
//
// See the method GetResponse and the function ReadBody to get more.
func (s Service) Do(method, bucket, object string, body, v interface{}, args ...Params) error {
	res, err := s.GetResponse(method, bucket, object, body, args...)
	if err != nil {
		return err
	}
	return ReadBody(res, v)
}

// just for test
var pause = 0 // seconds

// GetResponse sends an HTTP request to OSS and returns the HTTP response.
//
// The first optional Params is for Header, the second is for Query.
//
// See the method GetRequest to get more.
func (s Service) GetResponse(method, bucket, object string, body interface{}, args ...Params) (*http.Response, error) {
	if pause > 0 {
		time.Sleep(time.Duration(pause) * time.Second)
	}
	req, err := s.GetRequest(method, bucket, object, body, args...)
	if err != nil {
		return nil, err
	}
	s.Signature(req, 0)
	return http.DefaultClient.Do(req)
}

// GetRequest returns a new http.Request given a method and optional butcket, object, body.
//
// The first optional Params is for Header, the second is for Query.
//
// A nil body means no body, if the body's type is not
// []byte, *[]byte, *os.File, *bytes.Buffer, *bytes.Reader and *strings.Reader
// then encode XML as the body.
//
// The headers Content-Type and Content-Md5 will be set, when encode XML as the body
// or the body's type is []byte, *[]byte, *bytes.Buffer.
//
// It does not close the *os.File body.
//
// To signature the request call the method Signature.
func (s Service) GetRequest(method, bucket, object string, body interface{}, args ...Params) (*http.Request, error) {
	if s.AccessKeyId == "" || s.AccessKeySecret == "" {
		return nil, errAccessKeyRequired
	}

	header, query := getParams(args, 0).Header(), getParams(args, 1).Values()

	u := &url.URL{
		Scheme:   s.Scheme(),
		Host:     s.Host(),
		Path:     "/",
		RawQuery: query.Encode(),
	}

	if bucket != "" {
		if IsBucketName(bucket) {
			u.Host = bucket + "." + u.Host
		} else {
			return nil, errBucketNameInvalid
		}
	}

	if object != "" {
		if IsObjectName(object) {
			u.Path += object
		} else {
			return nil, errObjectNameInvalid
		}
	}

	if header.Get("User-Agent") == "" {
		header.Set("User-Agent", UserAgent)
	}
	if header.Get("Connection") == "" {
		header.Set("Connection", "keep-alive")
	}
	if header.Get("Accept-Encoding") == "" {
		header.Set("Accept-Encoding", "gzip, deflate")
	}
	if header.Get("Cache-Control") == "" {
		header.Set("Cache-Control", "no-cache")
	}

	req := &http.Request{
		Method:     method,
		URL:        u,
		Proto:      "HTTP/1.1", // 29
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
		Host:       u.Host,
	}

	setBody := func(v []byte) {
		req.Body = ioutil.NopCloser(bytes.NewReader(v))
	}

	setHeader := func(v []byte) {
		req.ContentLength = int64(len(v))
		req.Header.Set("Content-Type", http.DetectContentType(v))
		req.Header.Set("Content-Md5", Md5sum(v))
	}

	switch v := body.(type) {
	case nil:
	case []byte:
		setHeader(v)
		setBody(v)
	case *[]byte:
		setHeader(*v)
		setBody(*v)
	case *os.File:
		f, err := v.Stat()
		if err != nil {
			return nil, err
		}
		req.ContentLength = f.Size()
		req.Body = ioutil.NopCloser(v)
	case *bytes.Buffer:
		setHeader(v.Bytes())
		req.Body = ioutil.NopCloser(v)
	case *bytes.Reader:
		req.ContentLength = int64(v.Len())
		req.Body = ioutil.NopCloser(v)
	case *strings.Reader:
		req.ContentLength = int64(v.Len())
		req.Body = ioutil.NopCloser(v)
	default:
		b, err := xml.Marshal(body)
		if err != nil {
			return nil, err
		}
		setHeader(b)
		setBody(b)
	}

	return req, nil
}

// Signature url if the seconds > 0 and the Date Header add the seconds as the expires,
// otherwise signature and set the Authorization header.
//
// If the SecurityToken is not the empty string, then STS be supported.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/access-control&signature-header
// https://docs.aliyun.com/#/pub/oss/api-reference/access-control&signature-url
func (s Service) Signature(req *http.Request, seconds int) {
	header, u := req.Header, req.URL

	if s.SecurityToken != "" {
		if seconds > 0 {
			v := "security-token=" + s.SecurityToken
			if u.RawQuery != "" {
				v += "&"
			}
			u.RawQuery = v + u.RawQuery
		} else {
			header.Set("x-oss-security-token", s.SecurityToken)
		}
	}

	var expires string

	a := []string{req.Method, header.Get("Content-Md5"), header.Get("Content-Type")}

	if header.Get("Date") == "" {
		header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}
	if seconds > 0 {
		t, err := time.Parse(http.TimeFormat, header.Get("Date"))
		if err != nil {
			t = time.Now().UTC()
			header.Set("Date", t.Format(http.TimeFormat))
		}
		expires = strconv.FormatInt(t.Add(time.Duration(seconds)*time.Second).Unix(), 10)
		a = append(a, expires)
	} else {
		a = append(a, header.Get("Date"))
	}

	if v := CanonicalizedHeaders(req.Header); v != "" {
		a = append(a, v)
	}

	a = append(a, CanonicalizedResource(u))

	if v := HmacSha1(s.AccessKeySecret, strings.Join(a, "\n")); seconds > 0 {
		v = "OSSAccessKeyId=" + url.QueryEscape(s.AccessKeyId) +
			"&Expires=" + expires +
			"&Signature=" + url.QueryEscape(v)
		if u.RawQuery != "" {
			v += "&"
		}
		u.RawQuery = v + u.RawQuery
	} else {
		header.Set("Authorization", "OSS "+s.AccessKeyId+":"+v)
	}
}

// ListBucket returns all the buckets.
//
// The first optional Params is for Header, the second is for Query.
//
// Query predefine parameters: prefix, marker, max-keys.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/service&GetService
func (s Service) ListBucket(args ...Params) (*ListAllMyBucketsResult, error) {
	v := new(ListAllMyBucketsResult)

	err := s.Do("GET", "", "", nil, v, args...)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// ListAllMyBucketsResult represents the get service result.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/service&GetService
type ListAllMyBucketsResult struct {
	Prefix      string
	Marker      string
	MaxKeys     string
	IsTruncated bool
	NextMarker  string
	Owner       Owner
	Buckets     struct {
		Bucket []struct {
			Location     string
			Name         string
			CreationDate string
		}
	}
}
