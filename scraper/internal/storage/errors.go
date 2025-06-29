package storage

import (
	"errors"
)

var (
	ErrAlreadyExists = errors.New("data already exists")
	ErrNotExists     = errors.New("data does not exist")
)
