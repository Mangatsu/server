package utils

import (
	"regexp"
)

// Alphanumeric characters of any language, dashes, underscores, spaces, and special characters in the session name.
var wideRe = regexp.MustCompile(`^[\p{L}\p{N}\p{Pd}\p{Pc}\p{Zs}\p{Sc}\p{Sk}!?@#$%^&*+]+$`)

// The username can contain alphanumeric characters, dashes and underscores.
var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// The password can contain almost all characters except control characters, whitespace and quotes.
var passwordRe = regexp.MustCompile(`^[^\x00-\x1F\x7F\s'"]+$`)

const minUsernameLength = 2
const maxUsernameLength = 32
const minPasswordLength = 8
const maxPasswordLength = 512
const maxSessionNameLength = 128
const minCookieAge = 60
const maxCookieAge = 365 * 24 * 60 * 60 // year in seconds

// ClampCookieAge returns a valid cookie age in seconds.
func ClampCookieAge(seconds *int64) int64 {
	if seconds == nil {
		return maxCookieAge
	}

	return Clamp(*seconds, minCookieAge, maxCookieAge)
}

// IsValidSessionName checks if the session name is valid.
func IsValidSessionName(sessionName *string) bool {
	if sessionName == nil {
		return true
	}

	if !wideRe.MatchString(*sessionName) {
		return false
	}

	return len(*sessionName) <= maxSessionNameLength
}

// IsValidUsername checks if the username is valid.
func IsValidUsername(username string) bool {
	if len(username) < minUsernameLength || len(username) > maxUsernameLength {
		return false
	}

	return usernameRe.MatchString(username)
}

// IsValidPassword checks if the password is valid.
func IsValidPassword(password string) bool {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return false
	}

	return passwordRe.MatchString(password)
}
