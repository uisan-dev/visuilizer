package anilist

import "errors"

var ErrNotFound = errors.New("Not found")
var ErrAPIUnavailable = errors.New("anilist api is temporarily unavailable")
