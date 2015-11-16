// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"strconv"
	"time"
)

// InitiateMultipartUpload initialize a Multipart Upload event.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&InitiateMultipartUpload
func (o Object) InitiateMultipartUpload(args ...Params) (*InitiateMultipartUploadResult, error) {
	header, query := getHeaderQuery(args)
	query.Set("uploads", "")

	v := new(InitiateMultipartUploadResult)

	err := o.Do("POST", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// UploadPart upload the data as a part given a partNumber and a uploadId
// returns the ETag.
//
// The first optional Params is for Header, the second is for Query.
//
// The partNumber must be gte 1 and lte 10000.
// The data's type must be
// []byte, *[]byte, *os.File, *bytes.Buffer, *bytes.Reader or *strings.Reader.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&UploadPart
func (o Object) UploadPart(partNumber int, uploadId string, data interface{}, args ...Params) (string, error) {
	if !(partNumber >= 1 && partNumber <= 10000) {
		return "", errPartNumberInvalid
	}
	if uploadId == "" {
		return "", errUploadIdRequired
	}
	if !isPutDataType(data) {
		return "", errDataTypeNotSupported
	}

	header, query := getHeaderQuery(args)

	query.Set("partNumber", strconv.Itoa(partNumber))
	query.Set("uploadId", uploadId)

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

// UploadPartCopy upload a part copy from the source object given a partNumber and a uploadId
// returns the ETag.
//
// The first optional Params is for Header, the second is for Query.
//
// The partNumber must be gte 1 and lte 10000.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&UploadPartCopy
func (o Object) UploadPartCopy(partNumber int, uploadId string, source Object, args ...Params) (*CopyPartResult, error) {
	if !(partNumber >= 1 && partNumber <= 10000) {
		return nil, errPartNumberInvalid
	}
	if uploadId == "" {
		return nil, errUploadIdRequired
	}
	s := source.FullName()
	if s == "" {
		return nil, errSourceObjectInvalid
	}

	header, query := getHeaderQuery(args)

	header.Set("x-oss-copy-source", s)
	query.Set("partNumber", strconv.Itoa(partNumber))
	query.Set("uploadId", uploadId)

	v := new(CopyPartResult)

	err := o.Do("PUT", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// CompleteMultipartUpload complete the Multipart Upload given a uploadId, all the partNumbers and ETags.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&CompleteMultipartUpload
func (o Object) CompleteMultipartUpload(uploadId string, parts CompleteMultipartUpload, args ...Params) (*CompleteMultipartUploadResult, error) {
	if uploadId == "" {
		return nil, errUploadIdRequired
	}

	header, query := getHeaderQuery(args)
	query.Set("uploadId", uploadId)

	v := new(CompleteMultipartUploadResult)

	err := o.Do("POST", parts, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// AbortMultipartUpload abort the Multipart Upload given a uploadId,
// all the already uploaded parts will be deleted.
//
// To release all the OSS space call multiple times.
//
// The first optional Params is for Header, the second is for Query.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&AbortMultipartUpload
func (o Object) AbortMultipartUpload(uploadId string, args ...Params) error {
	if uploadId == "" {
		return errUploadIdRequired
	}

	header, query := getHeaderQuery(args)
	query.Set("uploadId", uploadId)

	return o.Do("DELETE", nil, nil, header, query)
}

// ListMultipartUploads returns the initialized but not complete or abort Multipart Uploads.
//
// The first optional Params is for Header, the second is for Query.
//
// Query predefine parameters: delimiter, max-uploads, key-marker, prefix, upload-id-marker, encoding-type.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&ListMultipartUploads
func (b Bucket) ListMultipartUploads(args ...Params) (*ListMultipartUploadsResult, error) {
	header, query := getHeaderQuery(args)
	query.Set("uploads", "")

	v := new(ListMultipartUploadsResult)

	err := b.Do("GET", "", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// ListParts returns the already uploaded parts given a uploadId.
//
// The first optional Params is for Header, the second is for Query.
//
// Query predefine parameters: max-parts, part-number-marker, encoding-type.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&ListParts
func (o Object) ListParts(uploadId string, args ...Params) (*ListPartsResult, error) {
	if uploadId == "" {
		return nil, errUploadIdRequired
	}

	header, query := getHeaderQuery(args)
	query.Set("uploadId", uploadId)

	v := new(ListPartsResult)

	err := o.Do("GET", nil, v, header, query)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// InitiateMultipartUploadResult represents the initialize Multipart Upload result.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&InitiateMultipartUpload
type InitiateMultipartUploadResult struct {
	Bucket   string
	Key      string
	UploadId string
}

// CopyPartResult represents the upload a copy part result.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&UploadPartCopy
type CopyPartResult struct {
	LastModified string
	ETag         string
}

// CompleteMultipartUploadPart represents a completed part.
type CompleteMultipartUploadPart struct {
	PartNumber int
	ETag       string
}

// CompleteMultipartUpload represents all the completed parts.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&CompleteMultipartUpload
type CompleteMultipartUpload struct {
	Part []CompleteMultipartUploadPart
}

// CompleteMultipartUploadResult represents the complete Multipart Upload result.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&CompleteMultipartUpload
type CompleteMultipartUploadResult struct {
	Bucket   string
	ETag     string
	Location string
	Key      string
}

// ListMultipartUploadsResult represents the list Multipart Uploads result.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&ListMultipartUploads
type ListMultipartUploadsResult struct {
	Bucket           string
	EncodingType     string
	KeyMarker        string
	UploadIdMarker   string
	NextKeyMarker    string
	NextUploadMarker string
	MaxUploads       int
	IsTruncated      bool
	Upload           []struct {
		Key       string
		UploadId  string
		Initiated time.Time
	}
}

// ListPartsResult represents the list parts result.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/api-reference/multipart-upload&ListParts
type ListPartsResult struct {
	Bucket               string
	EncodingType         string
	Key                  string
	UploadId             string
	PartNumberMarker     int
	NextPartNumberMarker string
	MaxParts             int
	IsTruncated          bool
	Part                 []struct {
		PartNumber   int
		LastModified time.Time
		ETag         string
		Size         int
	}
}
