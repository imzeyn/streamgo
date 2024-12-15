package streamgo

import (
	"errors"
	"regexp"
)

type HTTPMethod string

const (
	GET                HTTPMethod = "GET"
	POST               HTTPMethod = "POST"
	PUT                HTTPMethod = "PUT"
	DELETE             HTTPMethod = "DELETE"
	PATCH              HTTPMethod = "PATCH"
	OPTIONS            HTTPMethod = "OPTIONS"
	HEAD               HTTPMethod = "HEAD"
	TRACE              HTTPMethod = "TRACE"
	CONNECT            HTTPMethod = "CONNECT"	
)

var (
	singleBracePattern	  = regexp.MustCompile(`{.*?}`)
	doubleBracePattern	  = regexp.MustCompile(`{{.*?}}`)
)

var (
    ErrUploadExtMismatch 	   = errors.New("invalid file extension")
    ErrUploadSigMismatch  	   = errors.New("invalid signature")
    ErrUploadFileMissing  	   = errors.New("file not found")
)