package utils

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var telPattern = regexp.MustCompile(`^\+?[0-9]{9,15}$`)

type EmployeeInput struct {
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Birthday string `json:"birthday"`
	Email    string `json:"email"`
	Tel      string `json:"tel"`
}

func ValidateEmployee(in EmployeeInput, now time.Time) (time.Time, map[string]string) {
	errors := make(map[string]string)
	if !validName(in.Name) {
		errors["name"] = "required; use Thai or English letters and spaces only"
	}
	if !validName(in.Lastname) {
		errors["lastname"] = "required; use Thai or English letters and spaces only"
	}
	var birth time.Time
	if !datePattern.MatchString(in.Birthday) {
		errors["birthday"] = "required; use YYYY-MM-DD"
	} else {
		var err error
		birth, err = time.Parse("2006-01-02", in.Birthday)
		if err != nil || birth.Format("2006-01-02") != in.Birthday {
			errors["birthday"] = "must be a real calendar date"
		} else if birth.After(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)) {
			errors["birthday"] = "cannot be in the future"
		}
	}
	address, err := mail.ParseAddress(in.Email)
	if in.Email == "" || err != nil || address.Address != in.Email || !strings.Contains(in.Email, ".") {
		errors["email"] = "required; use a valid email address"
	}
	if !telPattern.MatchString(in.Tel) {
		errors["tel"] = "required; use 9 to 15 digits with optional leading +"
	}
	return birth, errors
}

func validName(s string) bool {
	if s == "" || s != strings.TrimSpace(s) {
		return false
	}
	seenLetter := false
	previousSpace := true
	for _, r := range s {
		switch {
		case r == ' ':
			if previousSpace {
				return false
			}
			previousSpace = true
		case isEnglishLetter(r) || (isThaiRune(r) && unicode.IsLetter(r)):
			seenLetter = true
			previousSpace = false
		case isThaiRune(r) && unicode.IsMark(r):
			if previousSpace {
				return false
			}
		default:
			return false
		}
	}
	return seenLetter && !previousSpace
}

func isEnglishLetter(r rune) bool { return r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' }

func isThaiRune(r rune) bool { return r >= '\u0E00' && r <= '\u0E7F' }
