package layout

import "errors"

var ErrCycle error = errors.New("relation graph contains a cycle")
