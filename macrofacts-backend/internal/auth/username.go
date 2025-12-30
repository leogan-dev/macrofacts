package auth

import (
	"errors"
	"regexp"
	"strings"
)

var (
	usernameRe = regexp.MustCompile(`^[A-Z0-9_.]{3,32}$`)

	ErrUsernameInvalid = errors.New("username_invalid")
	ErrUsernameTaken   = errors.New("username_taken")
	ErrPasswordInvalid = errors.New("password_invalid")
)

func CanonicalizeUsername(s string) (string, error) {
	u := strings.TrimSpace(s)
	u = strings.ToUpper(u)

	if !usernameRe.MatchString(u) {
		return "", ErrUsernameInvalid
	}
	return u, nil
}
