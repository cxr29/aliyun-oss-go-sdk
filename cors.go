// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"net/http"
)

// PutCORS set the CROS configuration, replace it if already exists.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/cors&PutBucketcors
func (b Bucket) PutCORS(cfg CORSConfiguration, args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("cors", "")
	return b.Do("PUT", "", cfg, nil, header, query)
}

// GetCORS get the CROS configuration.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/cors&GetBucketcors
func (b Bucket) GetCORS(args ...Params) (*CORSConfiguration, error) {
	header, query := getHeaderQuery(args)
	query.Set("cors", "")

	v := new(CORSConfiguration)

	err := b.Do("GET", "", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// DeleteCORS close the CORS and empty the rules.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/cors&DeleteBucketcors
func (b Bucket) DeleteCORS(args ...Params) error {
	header, query := getHeaderQuery(args)
	query.Set("cors", "")
	return b.Do("DELETE", "", nil, nil, header, query)
}

// Options the object and return the response header.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/cors&OptionObject
func (o Object) Options(args ...Params) (http.Header, error) {
	res, err := o.GetResponse("OPTIONS", nil, args...)
	if err != nil {
		return nil, err
	}

	err = ReadBody(res, nil)
	if err != nil {
		return nil, err
	}

	return res.Header, nil
}

// CORSRule represents a rule of the CORS configuration.
type CORSRule struct {
	AllowedOrigin string
	AllowedMethod string
	AllowedHeader string
	ExposeHeader  string
	MaxAgeSeconds int
}

// CORSConfiguration represents the CORS configuration.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/cors&PutBucketcors
// https://docs.aliyun.com/#/pub/oss/api-reference/cors&GetBucketcors
type CORSConfiguration struct {
	CORSRule []CORSRule
}
