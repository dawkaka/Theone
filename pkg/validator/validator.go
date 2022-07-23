package validator

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
)

func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsRealName(name string) bool {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 50 {
		return false
	}
	reg := regexp.MustCompile(`[0-9]|\W|_`)
	if reg.MatchString(name) {
		return false
	}
	if string(name[0]) != strings.ToUpper(string(name[0])) {
		return false
	}
	if name[1:] != strings.ToLower(name[1:]) {
		return false
	}
	return true
}

func IsUserName(userName string) bool {
	userName = strings.TrimSpace(userName)
	if len(userName) < 4 || len(userName) > 15 {
		return false
	}
	reg := regexp.MustCompile(`\W`)
	return !reg.MatchString(userName)
}

func IsCoupleName(coupleName string) bool {
	coupleName = strings.TrimSpace(coupleName)
	if len(coupleName) < 5 || len(coupleName) > 30 {
		return false
	}

	reg := regexp.MustCompile(`\w|\&`)
	return reg.MatchString(coupleName)
}

func IsPassword(password string) bool {
	password = strings.TrimSpace(password)
	return len(password) > 7
}

func IsBio(bio string) bool {
	return len(bio) < 256
}

func IsWebsite(website string) bool {
	return true
}

func IsPronouns(pronouns string) bool {
	return true
}

func IsValidPastDate(date time.Time) bool {
	return time.Now().After(date)
}

func IsCaption(caption string) bool {
	return len(caption) < 256
}
