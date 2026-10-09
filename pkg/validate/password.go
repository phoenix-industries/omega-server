package validate

import (
	"errors"
	"unicode"
)

func IsPassword(p string) bool {
	if len(p) < 8 || len(p) > 32 {
		return false
	}
	var number, upper bool
	for _, c := range p {
		if !number && unicode.IsNumber(c) {
			number = true
		}
		if !upper && unicode.IsUpper(c) {
			upper = true
		}
	}
	if !number || !upper {
		return false
	}
	return true
}

func Password(p string) error {
	if !IsPassword(p) {
		return errors.New("password must be between 8 and 32 characters (inclusive), and have at least 1 number and 1 upper case character")
	}
	return nil
}
