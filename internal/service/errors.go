package service

import "errors"

var (
	ErrInvalidCoordinaates = errors.New("lat or long is invalid")
	ErrTitleRequired       = errors.New("title is required")
	ErrInvalidRadius       = errors.New("radius must be more or equal than 5m")
)
