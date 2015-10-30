// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"hash"
	"regexp"
	"unicode/utf8"
)

var bucketNameRegexp = regexp.MustCompile(`^[-0-9a-z]{3,255}$`)

// IsBucketName returns true if the name is a valid bucket name.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/product-documentation/intruduction&concepts
func IsBucketName(name string) bool {
	return bucketNameRegexp.MatchString(name) &&
		name[0] != '-' &&
		name[len(name)-1] != '-'
}

// IsObjectName returns true if the name is a valid object name.
//
// Relevant documentation:
//
// https://docs.aliyun.com/#/pub/oss/product-documentation/intruduction&concepts
func IsObjectName(name string) bool {
	n := len(name)
	return n >= 1 && n <= 1023 &&
		name[0] != '/' && name[0] != '\\' &&
		utf8.ValidString(name)
}

// Md5sum returns the base64 MD5 checksum of the data.
func Md5sum(data []byte) string {
	a := md5.Sum(data)
	return base64.StdEncoding.EncodeToString(a[:])
}

// HmacSha1 returns the base64 HMAC-SHA1 hash of the data using the given secret.
func HmacSha1(secret, data string) string {
	return HmacX(secret, data, nil)
}

// HmacSha256 returns the base64 HMAC-SHA256 hash of the data using the given secret.
func HmacSha256(secret, data string) string {
	return HmacX(secret, data, sha256.New)
}

// HmacX returns the base64 HMAC-X hash of the data using the given secret.
//
// The SHA1 is used if the given hash.Hash type is nil.
func HmacX(secret, data string, h func() hash.Hash) string {
	if h == nil {
		h = sha1.New
	}
	x := hmac.New(h, []byte(secret))
	x.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(x.Sum(nil))
}
