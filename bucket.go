// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"net/http"
)

// Bucket represents a OSS bucket, the Name is required.
type Bucket struct {
	Service
	Name     string
	ACL      string
	Location string
}

// NewObject returns a new Object given a objectName from the bucket.
func (b Bucket) NewObject(objectName string) Object {
	return Object{
		Bucket: b,
		Name:   objectName,
	}
}

// Do sends an HTTP request to OSS and read the HTTP response to v.
//
// The first optional Params is for Header, the second is for Query.
//
// Overwritten the Service.Do with the bucket name.
func (b Bucket) Do(method, object string, body, v interface{}, args ...Params) error {
	if b.Name == "" {
		return errBucketNameRequired
	}
	return b.Service.Do(method, b.Name, object, body, v, args...)
}

// GetResponse sends an HTTP request to OSS and returns the HTTP response.
//
// The first optional Params is for Header, the second is for Query.
//
// Overwritten the Service.GetResponse with the bucket name.
func (b Bucket) GetResponse(method, object string, body interface{}, args ...Params) (*http.Response, error) {
	if b.Name == "" {
		return nil, errBucketNameRequired
	}
	return b.Service.GetResponse(method, b.Name, object, body, args...)
}

// GetRequest returns a new http.Request given a method and optional object, body.
//
// The first optional Params is for Header, the second is for Query.
//
// Overwritten the Service.GetRequest with the bucket name.
func (b Bucket) GetRequest(method, object string, body interface{}, args ...Params) (*http.Request, error) {
	if b.Name == "" {
		return nil, errBucketNameRequired
	}
	return b.Service.GetRequest(method, b.Name, object, body, args...)
}

// Put create the buckect also send the ACL and the location if they are not the empty string.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucket
func (b Bucket) Put(args ...Params) error {
	header, query := getHeaderQuery(args)
	if b.ACL != "" {
		header.Set("x-oss-acl", b.ACL)
	}
	var body interface{}
	if b.Location != "" {
		body = CreateBucketConfiguration{b.Location}
	}
	return b.Do("PUT", "", body, nil, header, query)
}

// PutACL change the bucket acl if it is the empty string.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucketACL
func (b Bucket) PutACL(args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("acl", "")
	if b.ACL != "" {
		header.Set("x-oss-acl", b.ACL)
	}
	return b.Do("PUT", "", nil, nil, header, query)
}

// PutLogging open the bucket access log.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucketLogging
func (b Bucket) PutLogging(bls BucketLoggingStatus, args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("logging", "")
	return b.Do("PUT", "", bls, nil, header, query)
}

// PutWebsite set the bucket as a static site.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucketWebsite
func (b Bucket) PutWebsite(wc WebsiteConfiguration, args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("website", "")
	return b.Do("PUT", "", wc, nil, header, query)
}

// PutReferer set the referer white list for preventing hotlinking.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucketReferer
// https://docs.aliyun.com/#/pub/oss/product-documentation/function&referer-white-list
func (b Bucket) PutReferer(rc RefererConfiguration, args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("referer", "")
	return b.Do("PUT", "", rc, nil, header, query)
}

// PutLifecycle set the objects lifecycle of the bucket.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucketLifecycle
func (b Bucket) PutLifecycle(lc LifecycleConfiguration, args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("lifecycle", "")
	return b.Do("PUT", "", lc, nil, header, query)
}

// ListObject returns the objects information of the bucket.
//
// The first optional Params is for Header, the second is for Query.
//
// Query predefine parameters: delimiter, marker, max-keys, prefix, encoding-type.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucket
func (b Bucket) ListObject(args ...Params) (*ListBucketResult, error) {
	v := new(ListBucketResult)

	err := b.Do("GET", "", nil, v, args...)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// GetACL returns the bucket ACL.
//
// The first optional Params is for Header, the second is for Query.
//
// Get and record:
//  b.ACL, err = b.GetACL()
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketAcl
func (b Bucket) GetACL(args ...Params) (string, error) {
	header, query := getHeaderQuery(args)
	query.Set("acl", "")
	var acp AccessControlPolicy
	err := b.Do("GET", "", nil, &acp, header, query)
	return acp.AccessControlList.Grant, err
}

// GetLocation returns the bucket location.
//
// The first optional Params is for Header, the second is for Query.
//
// Get and record:
//  b.Location, err = b.GetLocation()
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketLocation
func (b Bucket) GetLocation(args ...Params) (string, error) {
	header, query := getHeaderQuery(args)
	query.Set("location", "")
	var location string
	err := b.Do("GET", "", nil, &location, header, query)
	return location, err
}

// GetLogging returns the logging status of the bucket.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketLogging
func (b Bucket) GetLogging(args ...Params) (*BucketLoggingStatus, error) {
	header, query := getHeaderQuery(args)
	query.Set("logging", "")

	v := new(BucketLoggingStatus)

	err := b.Do("GET", "", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// GetWebsite returns the static site configuration of the bucket.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketWebsite
func (b Bucket) GetWebsite(args ...Params) (*WebsiteConfiguration, error) {
	header, query := getHeaderQuery(args)
	query.Set("website", "")

	v := new(WebsiteConfiguration)

	err := b.Do("GET", "", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// GetReferer returns the referer white list configuration of the bucket.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketReferer
func (b Bucket) GetReferer(args ...Params) (*RefererConfiguration, error) {
	header, query := getHeaderQuery(args)
	query.Set("referer", "")

	v := new(RefererConfiguration)

	err := b.Do("GET", "", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// GetLifecycle returns the objects lifecycle configuration of the bucket.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketLifecycle
func (b Bucket) GetLifecycle(args ...Params) (*LifecycleConfiguration, error) {
	header, query := getHeaderQuery(args)
	query.Set("lifecycle", "")

	v := new(LifecycleConfiguration)

	err := b.Do("GET", "", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// Delete the bucket, returns BucketNotEmpty Error if the buckect is not empty.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&DeleteBucket
func (b Bucket) Delete(args ...Params) error {
	return b.Do("DELETE", "", nil, nil, args...)
}

// DeleteLogging close the bucket access log.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&DeleteBucketLogging
func (b Bucket) DeleteLogging(args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("logging", "")
	return b.Do("DELETE", "", nil, nil, header, query)
}

// DeleteWebsite close the static site of the bucket.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&DeleteBucketWebsite
func (b Bucket) DeleteWebsite(args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("website", "")
	return b.Do("DELETE", "", nil, nil, header, query)
}

// DeleteLifecycle remove the objects lifecycle of the bucket, after no objects will be auto deleted.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&DeleteBucketLifecycle
func (b Bucket) DeleteLifecycle(args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("lifecycle", "")
	return b.Do("DELETE", "", nil, nil, header, query)
}

// DeleteObjects delete multiple objects by the keys, if not quiet returns the deleted objects.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&DeleteMultipleObjects
func (b Bucket) DeleteObjects(keys []string, quiet bool, args ...Params) ([]string, error) {
	header, query := getHeaderQuery(args)
	query.Set("delete", "")

	d := Delete{
		Object: make([]DeleteObject, len(keys)),
		Quiet:  quiet,
	}
	for k, v := range keys {
		d.Object[k].Key = v
	}

	var dr DeleteResult

	var v interface{}
	if !quiet {
		v = &dr
	}

	err := b.Do("POST", "", d, v, header, query)
	if err != nil {
		return nil, err
	}

	a := make([]string, len(dr.Deleted))
	for k, v := range dr.Deleted {
		a[k] = v.Key
	}

	return a, nil
}

// CreateBucketConfiguration represents the create bucket configuration.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucket
type CreateBucketConfiguration struct {
	LocationConstraint string
}

// BucketLoggingStatus represents the logging status.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucketLogging
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketLogging
type BucketLoggingStatus struct {
	LoggingEnabled struct {
		TargetBucket string
		TargetPrefix string
	}
}

// WebsiteConfiguration represents the static site configuration.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucketWebsite
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketWebsite
type WebsiteConfiguration struct {
	IndexDocument struct {
		Suffix string
	}
	ErrorDocument struct {
		Key string
	}
}

// RefererConfiguration represents the referer white list configuration.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucketReferer
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketReferer
type RefererConfiguration struct {
	AllowEmptyReferer bool
	RefererList       struct {
		Referer []string
	}
}

// LifecycleRule represents a rule of the objects lifecycle configuration.
type LifecycleRule struct {
	ID         string `xml:",omitempty"`
	Prefix     string
	Status     string // Enabled, Disabled
	Expiration struct {
		Date string `xml:",omitempty"`
		Days int    `xml:",omitempty"`
	}
}

// LifecycleConfiguration represents the objects lifecycle configuration.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&PutBucketLifecycle
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucketLifecycle
type LifecycleConfiguration struct {
	Rule []LifecycleRule
}

// ListBucketResult represents the get bucket result.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/bucket&GetBucket
type ListBucketResult struct {
	Name        string
	Prefix      string
	Marker      string
	MaxKeys     string
	Delimiter   string
	IsTruncated bool
	Contents    []struct {
		Key          string
		LastModified string
		ETag         string
		Type         string
		Size         string
		StorageClass string
		Owner        Owner
	}
	CommonPrefixes struct {
		Prefix string
	}
	NextMarker   string
	EncodingType string `xml:"encoding-type,omitempty"`
}

// DeleteObject represents a delete object.
type DeleteObject struct {
	Key string
}

// Delete represents the delete objects.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&DeleteMultipleObjects
type Delete struct {
	Quiet  bool
	Object []DeleteObject
}

// DeleteResult represents the delete objects result.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/object&DeleteMultipleObjects
type DeleteResult struct {
	Deleted []DeleteObject
}
