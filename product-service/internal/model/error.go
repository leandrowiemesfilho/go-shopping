package model

import "errors"

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidID       = errors.New("invalid product ID")
	ErrInvalidPrice    = errors.New("invalid price")
	ErrInvalidStock    = errors.New("invalid stock quantity")
	ErrDatabase        = errors.New("database error")
)
