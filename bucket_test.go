// Copyright 2015 Chen Xianren. All rights reserved.

package oss

import (
	"testing"
)

func TestBucket(t *testing.T) {
	b := newBucket()

	check := func(acl, location string) {
		var err error
		b.ACL, err = b.GetACL()
		fatal(t, err)

		equal(t, "acl", acl, b.ACL)

		b.Location, err = b.GetLocation()
		fatal(t, err)

		equal(t, "location", location, b.Location)
	}

	b.Service.Domain = ""
	fatal(t, b.Put())
	check(ACLPrivate, LocationCNHangzhou)
	fatal(t, b.Delete())

	b.Service.Domain = GetDomain(LocationCNShanghai, false)
	b.ACL = ACLPublicRead
	b.Location = LocationCNShanghai
	fatal(t, b.Put())
	check(ACLPublicRead, LocationCNShanghai)
	fatal(t, b.Delete())
}

func TestBucketACL(t *testing.T) {
	b := newBucket()

	fatal(t, b.Put())

	var err error
	b.ACL, err = b.GetACL()
	fatal(t, err)
	equal(t, "acl", ACLPrivate, b.ACL)

	b.ACL = "p"
	err = b.PutACL()
	e, ok := err.(Error)
	if !(ok && e.Code == "InvalidArgument") {
		t.Fatal("expected InvalidArgument")
	}

	b.ACL = ACLPublicRead
	fatal(t, b.PutACL())
	b.ACL = "r"
	b.ACL, err = b.GetACL()
	fatal(t, err)
	equal(t, "acl", ACLPublicRead, b.ACL)

	b.ACL = ACLPublicReadWrite
	fatal(t, b.PutACL())
	b.ACL = "rw"
	b.ACL, err = b.GetACL()
	fatal(t, err)
	equal(t, "acl", ACLPublicReadWrite, b.ACL)

	fatal(t, b.Delete())
}

func TestBucketLogging(t *testing.T) {
	b := newBucket()

	fatal(t, b.Put())

	v := new(BucketLoggingStatus)
	v.LoggingEnabled.TargetBucket = b.Name
	v.LoggingEnabled.TargetPrefix = "log-"

	fatal(t, b.PutLogging(*v))
	v = nil

	var err error
	v, err = b.GetLogging()
	fatal(t, err)
	equal(t, "TargetBucket", b.Name, v.LoggingEnabled.TargetBucket)
	equal(t, "TargetPrefix", "log-", v.LoggingEnabled.TargetPrefix)

	fatal(t, b.DeleteLogging())

	v, err = b.GetLogging()
	fatal(t, err)
	equal(t, "TargetBucket", "", v.LoggingEnabled.TargetBucket)
	equal(t, "TargetPrefix", "", v.LoggingEnabled.TargetPrefix)

	fatal(t, b.Delete())
}

func TestBucketWebsite(t *testing.T) {
	b := newBucket()

	fatal(t, b.Put())

	v := new(WebsiteConfiguration)
	v.IndexDocument.Suffix = "index.html"
	v.ErrorDocument.Key = "error.html"

	fatal(t, b.PutWebsite(*v))
	v = nil

	var err error
	v, err = b.GetWebsite()
	fatal(t, err)
	equal(t, "IndexDocument.Suffix", "index.html", v.IndexDocument.Suffix)
	equal(t, "ErrorDocument.Key", "error.html", v.ErrorDocument.Key)

	fatal(t, b.DeleteWebsite())

	v, err = b.GetWebsite()
	e, ok := err.(Error)
	if !(ok && e.Code == "NoSuchWebsiteConfiguration") {
		t.Fatal("expected NoSuchWebsiteConfiguration")
	}

	fatal(t, b.Delete())
}

func TestBucketReferer(t *testing.T) {
	b := newBucket()

	fatal(t, b.Put())

	referer := "http://www.cxr29.com"

	v := new(RefererConfiguration)
	v.AllowEmptyReferer = true
	v.RefererList.Referer = []string{referer}

	fatal(t, b.PutReferer(*v))
	v = nil

	var err error
	v, err = b.GetReferer()
	fatal(t, err)
	equal(t, "AllowEmptyReferer", true, v.AllowEmptyReferer)
	equal(t, "RefererList.Referer", 1, len(v.RefererList.Referer))
	equal(t, "RefererList.Referer", referer, v.RefererList.Referer[0])

	fatal(t, b.Delete())
}

func TestBucketLifecycle(t *testing.T) {
	b := newBucket()

	fatal(t, b.Put())

	v := new(LifecycleConfiguration)

	r := LifecycleRule{
		ID:     "delete after one day",
		Prefix: "logs",
		Status: "Enabled",
	}
	r.Expiration.Days = 1
	v.Rule = append(v.Rule, r)

	fatal(t, b.PutLifecycle(*v))
	v = nil

	var err error
	v, err = b.GetLifecycle()
	fatal(t, err)
	equal(t, "Rule", 1, len(v.Rule))
	equal(t, "Rule.ID", r.ID, v.Rule[0].ID)
	equal(t, "Rule.Prefix", r.Prefix, v.Rule[0].Prefix)
	equal(t, "Rule.Status", r.Status, v.Rule[0].Status)
	equal(t, "Rule.Expiration.Date", r.Expiration.Date, v.Rule[0].Expiration.Date)
	equal(t, "Rule.Expiration.Days", r.Expiration.Days, v.Rule[0].Expiration.Days)

	fatal(t, b.DeleteLifecycle())

	v, err = b.GetLifecycle()
	e, ok := err.(Error)
	if !(ok && e.Code == "NoSuchLifecycle") {
		t.Fatal("expected NoSuchLifecycle")
	}

	fatal(t, b.Delete())
}
