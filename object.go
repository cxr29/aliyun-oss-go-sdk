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
// If the bucket name or the object name is invalid returns the empty string.
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

// Range get the object range content to the data given the first-byte-pos and the length,
// returns the range-length and the instance-length.
//
// The range-length is less than or equal to the given length.
// The first-byte-pos add the range-length equal the instance-length indicate EOF.
//
// The first optional Params is for Header, the second is for Query.
//
// The data's type must be
// *[]byte, *os.File, *bytes.Buffer
//
// The last range must be given exact length or length <= 0,
// because OSS not support last-byte-pos greater than or equal to the instance-length.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&GetObject
func (o Object) Range(first, length int64, data interface{}, args ...Params) (int64, int64, error) {
	if !isGetDataType(data) {
		return 0, 0, errDataTypeNotSupported
	}

	r := FormatRange(first, length)
	if r == "" {
		return 0, 0, errRangeInvalid
	}

	header, query := getHeaderQuery(args)
	header.Set(HeaderRange, r)

	res, err := o.GetResponse("GET", nil, header, query)
	if err != nil {
		return 0, 0, err
	}

	err = newBody(res)
	defer res.Body.Close()
	if err == nil {
		err = readError(res)
	}
	if err != nil {
		return 0, 0, err
	}

	f, l, t, err := ParseContentRange(res.Header.Get(HeaderContentRange))
	if err != nil {
		return 0, 0, err
	}

	if f == -1 || l == -1 || t == -1 || f != first || (length > 0 && l > length) {
		return 0, 0, ErrContentRangeCorrupt
	}

	if n, err := readBody(res.Body, data); err != nil {
		return 0, 0, err
	} else if n != l {
		return 0, 0, ErrContentRangeCorrupt
	}

	return l, t, nil
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

// Head the object and returns the response header.
//
// The first optional Params is for Header, the second is for Query.
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

// GetInfo returns the object info.
//
// The first optional Params is for Header, the second is for Query.
func (o Object) GetInfo(args ...Params) (*GetObjectInfoResult, error) {
	header, query := getHeaderQuery(args)

	query.Set("objectInfo", "")

	v := new(GetObjectInfoResult)

	err := o.Do("GET", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// GetObjectInfoResult represents the object info result.
type GetObjectInfoResult struct {
	Bucket       string
	Type         string
	Key          string
	ETag         string
	ContentType  string `xml:"Content-Type"`
	Size         int64
	LastModified string // 2006-01-02T15:04:05.000Z
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

// GetACL returns the object ACL, if not set returns "default".
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
