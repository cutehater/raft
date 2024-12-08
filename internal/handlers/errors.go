package handlers

import "errors"

var (
	errInvalidID      = errors.New("invalid resource Id")
	errNotFound       = errors.New("resource not found")
	errInvalidMethod  = errors.New("invalid HTTP method")
	errInvalidData    = errors.New("invalid resource data")
	errResourceExists = errors.New("resource already exists")
)
