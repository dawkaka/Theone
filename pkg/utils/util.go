package utils

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/gin-contrib/sessions"
)

//User's prefered language for success or error messages
func GetLang(userLang string, header http.Header) string {
	if userLang != "" {
		return userLang
	}
	lang := header.Get("Accept-Language")

	if lang != "" {
		l := strings.Split(lang, "-")[0]
		if validator.IsSupportedLanguage(l) {
			return l
		}
	}
	return "en"
}

//GenerateId short ids for couple post
//to make links shared short and nice!
//Mongodb's default ids are long and ugly
func GenerateID() string {
	alphabets := "abc8debg7hijkl0mn6GH5IJKLMNo9pq1rstuv2wxy3zABCD4EFOPQRSTUVWSYZ"
	var id string
	for i := 0; i < 12; i++ {
		rand.Seed(time.Now().UnixNano())
		ind := rand.Intn(61)
		id += string(alphabets[ind])
	}
	return id
}

//Get session data

func GetSession(session sessions.Session) entity.UserSession {
	var userSession entity.UserSession
	val := session.Get("user")
	if val != nil {
		userSession = val.(entity.UserSession)
	}
	return userSession
}

//ExtractMentions extracts all users mention (@) so that they can be notified
func ExtracMentions(caption string) []string {
	mentions := []string{}
	if len(strings.TrimSpace(caption)) == 0 {
		return mentions
	}
	captionWords := strings.Split(caption, " ")
	for _, val := range captionWords {
		val = strings.TrimSpace(val)
		if val[0] == '@' {
			userName := val[1:]
			if validator.IsUserName(userName) {
				mentions = append(mentions, userName)
			}
		}
	}
	return mentions
}

func GetCategory(level int, country string) []string {

	var target string
	for key, val := range one {
		for _, v := range val {
			if v == country {
				target = key
			}
		}
	}
	if level == 1 {
		return one[target]
	}

	for key, val := range two {
		for _, v := range val {
			if v == target {
				target = key
			}
		}
	}
	res := []string{}
	for _, v := range two[target] {
		res = append(res, one[v]...)
	}
	return res
}

var one = map[string][]string{
	"AfBEn": {"Ghana", "Nigeria", "The Gambia", "Sierra Leone", "Liberia", "Kenya", "Uganda", "Tanzania", "South Africa"},
	"AfBFr": {"Senegal", "Guinea-Bissau", "Guinea", "Ivory Coast", "Mali", "Togo", "Benin", "Burkina Faso", "Niger"},
	"AfAr":  {"Egypt", "Algeria", "Morrocco", "Libya", "Sudan", "Tunisia", "Chad", "Djibouti", "Comoros"},
	"AfPt":  {"Angola", "Cape Verde", "Guinea-Bissau", "Mozambique", "São Tomé", "Príncipe"},
	"MeAr": {
		"Bahrain", "United Arab Emirates", "Jordan", "Iraq", "Qatar",
		"Saudi Arabia", "Oman", "Syria", "Kuwait", "Qatar", "Israel", "Lebanon",
	},
	"SAm_es": {
		"Argentina", "Dominica",
		"Columbia", "Ecuador", "Paraguay", "Peru", "Uruguay",
		"Bolivia", "Chile", "Colombia", "Venezuela", "Cuba",
	},
	"Es_Pt": {"Portugal", "Brazil"},
	"NAm_es": {
		"Mexico", "Guatemala", "Honduras", "El Salvador", "Nicaragua", "Panama",
		"Guatemela", "Dominican Republic", "Costa Rica",
	},
	"Eu_es": {
		"Spain", "Puerto Rico",
	},
	"Eu_fr": {
		"France", "Belgium",
	},
	"AmWEn":  {"United States of America", "United Kingdom", "Australia", "Canada", "New Zealand", "Ireland"},
	"AmWBEn": {"The Bahamas", "Barbados", "Belize", "Jamaica", "Dominica", "Grenada", "Guyana", "Antigua and Barbuda"},
	"Ge":     {"Germany", "Austria", "Switzerland", "Luxembourg"},
}

var two = map[string][]string{
	"AfBl":       {"AfBEn", "AfBFr"},
	"Eu_mix":     {"Eu_fr", "Ge"},
	"America_en": {"AmWEn", "AmWBen"},
	"Ar":         {"AfAr", "MeAr"},
	"Es":         {"SAm_es", "NAm_es", "Eu_es"},
	"Pt":         {"Es_Pt", "AfPt"},
}
