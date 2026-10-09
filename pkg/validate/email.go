package validate

import (
	"errors"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func IsEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func Email(email string) error {
	if !IsEmail(email) {
		return errors.New("invalid email address")
	}
	return nil
}
