package utils

import (
	"errors"
	"strings"
)

func FormatDBError(err string) error {
	if strings.Contains(err, "nickname") {
		return errors.New("nickname already taken")
	}
	if strings.Contains(err, "email") {
		return errors.New("email already taken")
	}
	return errors.New("incorrect details")
}
