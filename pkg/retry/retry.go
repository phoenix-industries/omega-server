package retry

import "errors"

const MaxRetries = 5

var errMaxRetries = errors.New("exceeded retry limit")

type Func func(attempt int) (retry bool, err error)

func Do(fn Func) error {
	var con bool
	var err error
	for att := 1; att <= MaxRetries; att++ {
		con, err = fn(att)
		if !con || err == nil {
			return nil
		}
	}
	return errMaxRetries
}

func IsMaxRetriesError(err error) bool {
	return errors.Is(err, errMaxRetries)
}
