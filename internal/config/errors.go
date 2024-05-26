package config

import "errors"

var (
	ErrConfigLoadFail    = errors.New("failed to load configuration file")
	ErrConfigBindingFail = errors.New("failed to binding configuration")
)
