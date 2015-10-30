// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"net/http"
	"strconv"
)

// Object represents a OSS object, the Name is required.
type Object struct {
	Bucket
	Name string
	ACL  string
}

// FullName returns the string "/BucketName/ObjectName".
//
// If the bucket name or the object name is invalid return the empty string.
func (o Object) FullName() string {
	if IsBucketName(o.Bucket.Name) && IsObjectName(o.Name) {
		return "/" + o.Bucket.Name + "/" + o.Name
	}
	return ""
}

// Do sends an HTTP request to OSS and read the HTTP response to v.
//
// The first optional Params is for Header, the second is for Query.
//
// Overwritten the Bucket.Do with the object name.
func (o Object) Do(method string, body, v interface{}, args ...Params) error {
	if o.Name == "" {
		return errObjectNameRequired
	}
	return o.Bucket.Do(method, o.Name, body, v, args...)
}

// GetResponse sends an HTTP request to OSS and returns the HTTP response.
//
// The first optional Params is for Header, the second is for Query.
//
// Overwritten the Bucket.GetResponse with the object name.
func (o Object) GetResponse(method string, body interface{}, args ...Params) (*http.Response, error) {
	if o.Name == "" {
		return nil, errObjectNameRequired
	}
	return o.Bucket.GetResponse(method, o.Name, body, args...)
}

// GetRequest returns a new http.Request given a method and body.
//
// The first optional Params is for Header, the second is for Query.
//
// Overwritten the Bucket.GetRequest with the object name.
func (o Object) GetRequest(method string, body interface{}, args ...Params) (*http.Request, error) {
	if o.Name == "" {
		return nil, errObjectNameRequired
	}
	return o.Bucket.GetRequest(method, o.Name, body, args...)
}

// Put the data as the object content,
// also send the ACL if it is not the empty string,
// returns the ETag.
//
// The first optional Params is for Header, the second is for Query.
//
// The data's type must be
// []byte, *[]byte, *os.File, *bytes.Buffer, *bytes.Reader or *strings.Reader.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&PutObject
func (o Object) Put(data interface{}, args ...Params) (string, error) {
	if !isPutDataType(data) {
		return "", errDataTypeNotSupported
	}

	header, query := getHeaderQuery(args)
	if o.ACL != "" {
		header.Set("x-oss-object-acl", o.ACL)
	}

	res, err := o.GetResponse("PUT", data, header, query)
	if err != nil {
		return "", err
	}

	err = ReadBody(res, nil)
	if err != nil {
		return "", err
	}

	return res.Header.Get("ETag"), nil
}

// Copy the object content from the source object,
// also send the ACL if it is not the empty string.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&CopyObject
func (o Object) Copy(source Object, args ...Params) (*CopyObjectResult, error) {
	s := source.FullName()
	if s == "" {
		return nil, errSourceObjectInvalid
	}

	header, query := getHeaderQuery(args)
	header.Set("x-oss-copy-source", s)

	if o.ACL != "" {
		header.Set("x-oss-object-acl", o.ACL)
	}

	v := new(CopyObjectResult)

	err := o.Do("PUT", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// Get the object content to the data.
//
// The first optional Params is for Header, the second is for Query.
//
// The data's type must be
// *[]byte, *os.File, *bytes.Buffer
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&GetObject
func (o Object) Get(data interface{}, args ...Params) error {
	if !isGetDataType(data) {
		return errDataTypeNotSupported
	}
	return o.Do("GET", nil, data, args...)
}

// Append the data to the object content, also send the ACL if it is not the empty string,
// returns ObjectNotAppendable Error if the object is not Appendable,
// returns the next append position, the hash crc64ecma and the ETag if success.
//
// The first optional Params is for Header, the second is for Query.
//
// The data's type must be
// []byte, *[]byte, *os.File, *bytes.Buffer, *bytes.Reader or *strings.Reader.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&AppendObject
func (o Object) Append(position int64, data interface{}, args ...Params) (next int64, crc, etag string, err error) {
	if !isPutDataType(data) {
		err = errDataTypeNotSupported
		return
	}

	header, query := getHeaderQuery(args)

	query.Set("append", "")
	query.Set("position", strconv.FormatInt(position, 10))

	if o.ACL != "" {
		header.Set("x-oss-object-acl", o.ACL)
	}

	res, err := o.GetResponse("POST", data, header, query)
	if err != nil {
		return
	}

	err = ReadBody(res, nil)
	if err != nil {
		return
	}

	next, err = strconv.ParseInt(res.Header.Get("x-oss-next-append-position"), 10, 64)
	crc, etag = res.Header.Get("x-oss-hash-crc64ecma"), res.Header.Get("ETag")
	return
}

// Delete the object.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&DeleteObject
func (o Object) Delete(args ...Params) error {
	return o.Do("DELETE", nil, nil, args...)
}

// Head the object returns the meta.
//
// The first optional Params is for Header, the second is for Query.
//
// Head the object and return the response header.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&HeadObject
func (o Object) Head(args ...Params) (http.Header, error) {
	res, err := o.GetResponse("HEAD", nil, args...)
	if err != nil {
		return nil, err
	}
	err = ReadBody(res, nil)
	if err != nil {
		return nil, err
	}
	return res.Header, nil
}

// PutACL change the object acl if it is the empty string.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&PutObjectACL
func (o Object) PutACL(args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("acl", "")
	if o.ACL != "" {
		header.Set("x-oss-object-acl", o.ACL)
	}
	return o.Do("PUT", nil, nil, header, query)
}

// GetACL returns the object ACL, if not set return "default".
//
// The first optional Params is for Header, the second is for Query.
//
// Get and record:
//  o.ACL, err = o.GetACL()
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&GetObjectACL
func (o Object) GetACL(args ...Params) (string, error) {
	header, query := getHeaderQuery(args)
	query.Set("acl", "")
	var acp AccessControlPolicy
	err := o.Do("GET", nil, &acp, header, query)
	return acp.AccessControlList.Grant, err
}

// CopyObjectResult represents the copy object result.
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&CopyObject
type CopyObjectResult struct {
	LastModified string
	ETag         string
}
