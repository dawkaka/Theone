package validator

import (
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	SUPPORTED_LANGUAGES = []string{"en", "es", "ch", "fr", "ar", "ru"}
	Settings            = map[string][]string{"language": SUPPORTED_LANGUAGES}
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
	if isASCII(name) {

		for _, val := range name {
			if !unicode.IsLetter(val) && val != '\'' {
				return false
			}
		}
		if string(name[0]) != strings.ToUpper(string(name[0])) {
			return false
		}
		if name[1:] != strings.ToLower(name[1:]) {
			return false
		}
		return true
	}
	return true
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func IsUserName(userName string) bool {
	userName = strings.TrimSpace(userName)
	nameLength := len(userName)
	if nameLength < 4 || nameLength > 15 {
		return false
	}
	reg := regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
	if !reg.MatchString(userName) {
		return false
	}
	//User name should not start or end with a special char
	if userName[0] == '_' || userName[0] == '.' || userName[nameLength-1] == '_' || userName[nameLength-1] == '.' {
		return false
	}
	for i := 0; i < len(userName)-1; i++ {
		if userName[i] == '_' || userName[i] == '.' {
			if userName[i+1] == '_' || userName[i+1] == '.' {
				return false
			}
		}
	}
	return true
}

func IsCoupleName(coupleName string) bool {
	coupleName = strings.TrimSpace(coupleName)
	nameLength := len(coupleName)
	if nameLength < 5 || nameLength > 30 {
		return false
	}
	reg := regexp.MustCompile(`^[a-zA-Z0-9_.&]+$`)
	if !reg.MatchString(coupleName) {
		return false
	}
	spec := "._&"
	if strings.Contains(spec, string(coupleName[0])) || strings.Contains(spec, string(coupleName[nameLength-1])) {
		return false
	}
	for i := 0; i < len(coupleName)-1; i++ {
		if coupleName[i] == '_' || coupleName[i] == '.' {
			if coupleName[i+1] == '_' || coupleName[i+1] == '.' {
				return false
			}
		}
	}
	return true
}

func IsPassword(password string) bool {
	password = strings.TrimSpace(password)
	return len(password) > 7
}

func IsBio(bio string) bool {
	return len(bio) < 256
}

func IsWebsite(website string) bool {
	return len(website) < 100
}

func IsPronouns(pronouns string) bool {
	pronouns = strings.TrimSpace(pronouns)
	ps := strings.Split(pronouns, "/")
	l := len(ps)
	for _, val := range ps {
		if val == "" {
			return false
		}
	}
	return l < 3 && l > 1
}

func IsValidPastDate(date time.Time) bool {
	return time.Now().After(date)
}

func IsCaption(caption string) bool {
	return len(caption) < 256
}

func IsSupportedImageType(content []byte) (imageType string, supported bool) {
	imageType = http.DetectContentType(content)
	for _, val := range []string{"image/gif", "image/jpeg", "image/jpg", "image/png"} {
		if val == imageType {
			supported = true
			return
		}
	}
	supported = false
	return
}

func IsValidSetting(setting, value string) bool {

	if settings, ok := Settings[setting]; ok {
		for _, val := range settings {
			if val == value {
				return true
			}
		}
	}
	return false
}
