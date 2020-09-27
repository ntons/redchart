package ranking

import (
	"errors"
)

var (
	ErrNameExist    = errors.New("name exist")
	ErrNameNotExist = errors.New("name not exist")
)
