package oss

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//
const (
	HeaderRange        = "Range"
	HeaderContentRange = "Content-Range"
)

//
var (
	ErrContentRangeInvalid = errors.New("content range invalid")
	ErrContentRangeCorrupt = errors.New("content range corrupt")
)

var reContentRangePrefix = regexp.MustCompile(`(?i)^bytes\s+`)

// FormatRange set the Range header given first and length.
// It returns byte-range-spec if first >= 0, no last-byte-pos if length <= 0.
// It returns suffix-byte-range-spec if first < 0 and length > 0.
// It returns the empty string if first < 0 and length <= 0.
//
// http://tools.ietf.org/html/draft-ietf-httpbis-p5-range-19#section-5.4
func FormatRange(first, length int64) (r string) {
	if first < 0 {
		if length <= 0 {
			return ""
		}
		r = fmt.Sprintf("-%d", length)
	} else {
		if length <= 0 {
			r = fmt.Sprintf("%d-", first)
		} else {
			r = fmt.Sprintf("%d-%d", first, first+length-1)
		}
	}
	return "bytes=" + r
}

// ParseContentRange parse the Content-Range header.
// It returns total = -1 if the instance-length is the asterisk "*" character.
// It returns first = length = -1 if the byte-range-resp-spec is the asterisk "*" character.
//
// http://tools.ietf.org/html/draft-ietf-httpbis-p5-range-19#section-5.2
func ParseContentRange(cr string) (first int64, length int64, total int64, err error) {
	loc := reContentRangePrefix.FindStringIndex(strings.TrimSpace(cr))
	if loc == nil {
		return 0, 0, 0, ErrContentRangeInvalid
	}
	cr = cr[loc[1]:]

	if strings.HasPrefix(cr, "*/") { // 416
		total, err = strconv.ParseInt(cr[2:], 10, 64)
		if err != nil || total < 0 {
			return 0, 0, 0, ErrContentRangeInvalid
		}
		return -1, -1, total, nil
	}

	var idx int

	if strings.HasSuffix(cr, "/*") {
		total = -1
		cr = cr[:len(cr)-2]
	} else {
		idx = strings.LastIndex(cr, "/")
		if idx == -1 {
			return 0, 0, 0, ErrContentRangeInvalid
		}

		total, err = strconv.ParseInt(cr[idx+1:], 10, 64)
		if err != nil || total < 0 {
			return 0, 0, 0, ErrContentRangeInvalid
		}

		cr = cr[:idx]
	}

	idx = strings.Index(cr, "-")
	if idx == -1 {
		return 0, 0, 0, ErrContentRangeInvalid
	}

	first, err = strconv.ParseInt(cr[:idx], 10, 64)
	if err != nil || first < 0 {
		return 0, 0, 0, ErrContentRangeInvalid
	}

	length, err = strconv.ParseInt(cr[idx+1:], 10, 64)
	if err != nil || length < 0 {
		return 0, 0, 0, ErrContentRangeInvalid
	}

	if length >= first && (total == -1 || total > length) {
		return first, length - first + 1, total, nil
	}

	return 0, 0, 0, ErrContentRangeCorrupt
}
