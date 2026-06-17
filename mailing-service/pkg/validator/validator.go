package validator

import (
	"net/mail"
	"strings"
)

func IsValidEmail(addr string) bool {
	_, err := mail.ParseAddress(addr)
	return err == nil
}

func IsValidEmailList(addresses []string) ([]string, bool) {
	var invalid []string
	for _, a := range addresses {
		if !IsValidEmail(strings.TrimSpace(a)) {
			invalid = append(invalid, a)
		}
	}
	return invalid, len(invalid) == 0
}
