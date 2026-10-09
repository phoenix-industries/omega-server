package validate

import (
	"errors"

	"github.com/nyaruka/phonenumbers"
)

func IsPhoneNumber(phone string) bool {
	_, err := phonenumbers.Parse(phone, "EG")
	return err == nil
}

func PhoneNumber(phone string) error {
	if !IsPhoneNumber(phone) {
		return errors.New("invalid phone number")
	}
	return nil
}
