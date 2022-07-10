package validator

import (
	"net/mail"
	"strings"
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
	if strings.ContainsAny(name, ".#@/][*)(&^%$!~`\"\\{}<>,_+=1234567890|?") {
		return false
	}
	if string(name[0]) != strings.ToUpper(string(name[0])) {
		return false
	}
	if name[1:] != strings.ToLower(name) {
		return false
	}

	return true
}

func IsUserName(userName string) bool {
	userName = strings.TrimSpace(userName)
	if len(userName) < 4 || len(userName) > 15 {
		return false
	}

	if strings.ContainsAny(userName, ".#@/][*)(&^%$!~`'\"\\{}<>,-+=|?") {
		return false
	}

	return true
}

func IsCoupleName(coupleName string) bool {
	coupleName = strings.TrimSpace(coupleName)
	if len(coupleName) < 5 || len(coupleName) > 30 {
		return false
	}

	if strings.ContainsAny(coupleName, ".#@/][*)(^%$!~`'\"\\{}<>,-+=|?") {
		return false
	}

	return true
}

func IsPassword(password string) bool {
	password = strings.TrimSpace(password)
	return len(password) > 7
}
