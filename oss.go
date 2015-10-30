// Copyright 2015 Chen Xianren. All rights reserved.

// Package oss implements a library for Aliyun Object Storage Service.
package oss

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

// OSS Access Control List
const (
	ACLPublicReadWrite = "public-read-write"
	ACLPublicRead      = "public-read"
	ACLPrivate         = "private"
)

// OSS Location List
const (
	LocationCNQingdao    = "oss-cn-qingdao"
	LocationCNBeijing    = "oss-cn-beijing"
	LocationCNHangzhou   = "oss-cn-hangzhou"
	LocationCNHongkong   = "oss-cn-hongkong"
	LocationCNShenzhen   = "oss-cn-shenzhen"
	LocationCNShanghai   = "oss-cn-shanghai"
	LocationUSWest1      = "oss-us-west-1"
	LocationAPSoutheast1 = "oss-ap-southeast-1"
)

// UserAgent is the default user agent and is used by GetRequest.
var UserAgent = "aliyun-oss-go-sdk"

// GetDomain returns the OSS access domain by the location,
// if the internal is ture returns the intranet domain.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/product-documentation/domain-region
func GetDomain(location string, internal bool) (domain string) {
	if location != "" {
		domain = location
	} else {
		domain = "oss"
	}
	if internal {
		domain += "-internal"
	}
	return domain + ".aliyuncs.com"
}

// ReadBody reads the response body and stores the result in the value pointed to by v,
//
// When the status code is 3xx, 4xx or 5xx return the Error.
//
// If v is nil, the response body is discarded or, if v's type is not
// *[]byte, *os.File and *bytes.Buffer,
// then decode XML to it.
//
// Content-Encoding deflate and gzip are supported.
func ReadBody(res *http.Response, v interface{}) error {
	defer res.Body.Close()

	var rc io.ReadCloser
	var err error

	switch res.Header.Get("Content-Encoding") {
	case "deflate":
		rc = flate.NewReader(res.Body)
		defer rc.Close()
	case "gzip":
		rc, err = gzip.NewReader(res.Body)
		if err != nil {
			return err
		}
		defer rc.Close()
	default:
		rc = res.Body
	}

	switch res.StatusCode / 100 {
	case 3, 4, 5: // 7
		b, err := ioutil.ReadAll(rc)
		if err != nil {
			return err
		}
		var e Error
		if len(b) > 0 { // no content
			err = xml.Unmarshal(b, &e)
			if err != nil {
				return err
			}
		} else {
			e.Code = res.Status
		}
		return e
	}

	switch i := v.(type) {
	case nil:
	case *[]byte:
		*i, err = ioutil.ReadAll(rc)
	case *os.File:
		_, err = io.Copy(i, rc)
	case *bytes.Buffer:
		_, err = i.ReadFrom(rc)
	default:
		err = xml.NewDecoder(rc).Decode(v)
	}

	return err
}

type dict [][2]string

func (a dict) Len() int {
	return len(a)
}
func (a dict) Less(i, j int) bool {
	return a[i][0] < a[j][0]
}
func (a dict) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a dict) Sort() {
	sort.Sort(a)
}

// CanonicalizedHeaders returns the canonicalized OSS headers as a string.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/access-control&signature-header
func CanonicalizedHeaders(header http.Header) string {
	if len(header) > 0 {
		var a dict
		for k, v := range header {
			if l := strings.ToLower(k); strings.HasPrefix(l, "x-oss-") {
				x := ""
				if len(v) > 0 {
					x = v[0]
				}
				a = append(a, [2]string{l, x})
			}
		}
		if n := len(a); n > 0 {
			a.Sort()
			b := make([]string, n)
			for k, v := range a {
				b[k] = v[0] + ":" + v[1]
			}
			return strings.Join(b, "\n")
		}
	}
	return ""
}

var resources = []string{
	"acl",
	"append",
	"cors",
	"delete",
	"group",
	"lifecycle",
	"link",
	"location",
	"logging",
	"objectInfo",
	"partNumber",
	"position",
	"qos",
	"referer",
	"response-cache-control",
	"response-content-disposition",
	"response-content-encoding",
	"response-content-language",
	"response-content-type",
	"response-expires",
	"restore",
	"security-token",
	"uploadId",
	"uploads",
	"website",
}

// CanonicalizedResource returns the canonicalized OSS resource as a string.
//
// Get the bucket name and the object name from the URL's Host and Path.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/access-control&signature-header
func CanonicalizedResource(u *url.URL) string {
	if u == nil {
		return ""
	}
	s := "/"
	if a := strings.Split(u.Host, "."); len(a) == 4 {
		s += a[0] + "/"
	}
	if u.Path != "" && u.Path[0] == '/' {
		s += u.Path[1:]
	} else {
		s += u.Path
	}
	if q := u.Query(); len(q) > 0 {
		var a dict
		for _, k := range resources {
			if v, ok := q[k]; ok {
				x := ""
				if len(v) > 0 {
					x = v[0]
				}
				a = append(a, [2]string{k, x})
			}
		}
		if n := len(a); n > 0 {
			a.Sort()
			b := make([]string, n)
			for k, v := range a {
				b[k] = v[0]
				if v[1] != "" {
					b[k] += "=" + v[1]
				}
			}
			s += "?" + strings.Join(b, "&")
		}
	}
	return s
}

// A Params represents the http.Header or the url.Values.
type Params map[string][]string

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map directly.
func (args Params) Get(key string) string {
	if args == nil {
		return ""
	}
	value, ok := args[key]
	if !ok || len(value) == 0 {
		return ""
	}
	return value[0]
}

// Set sets the key to value. It replaces any existing values.
func (args Params) Set(key, value string) {
	args[key] = []string{value}
}

// Add adds the value to key. It appends to any existing values associated with key.
func (args Params) Add(key, value string) {
	args[key] = append(args[key], value)
}

// Del deletes the values associated with key.
func (args Params) Del(key string) {
	delete(args, key)
}

// Copy all the params's keys and values to args.
func (args Params) Copy(params Params) {
	for k, v := range params {
		args[k] = make([]string, len(v))
		copy(args[k], v)
	}
}

// Values convert Params to url.Values.
func (args Params) Values() url.Values {
	return url.Values(args)
}

// Header convert Params to http.Header.
func (args Params) Header() http.Header {
	return http.Header(args)
}

func getParams(args []Params, index int) Params {
	if len(args) > index && args[index] != nil {
		return args[index]
	}
	return Params{}
}

func getHeaderQuery(args []Params) (Params, Params) {
	return getParams(args, 0), getParams(args, 1)
}

func isPutDataType(data interface{}) bool {
	switch data.(type) {
	case []byte, *[]byte, *os.File, *bytes.Buffer, *bytes.Reader, *strings.Reader:
		return true
	default:
		return false
	}
}

func isGetDataType(data interface{}) bool {
	switch data.(type) {
	case *[]byte, *os.File, *bytes.Buffer:
		return true
	default:
		return false
	}
}

// Error represents the OSS error response when the status code is 3xx, 4xx or 5xx.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/error-response
type Error struct {
	Code, Message string
	RequestId     string
	HostId        string
}

func (e Error) Error() string {
	return fmt.Sprintf("Code: %s, Message: %s, RequestId: %s, HostId: %s", e.Code, e.Message, e.RequestId, e.HostId)
}

// Owner contains the information of the bucket owner.
type Owner struct {
	ID          string
	DisplayName string
}

// AccessControlPolicy contains the information of the bucket ACL.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/access-control&bucket-acl
type AccessControlPolicy struct {
	Owner             Owner
	AccessControlList struct {
		Grant string
	}
}
