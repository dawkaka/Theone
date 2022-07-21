package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/dawkaka/theone/pkg/validator"
)

//User's prefered language for success or error messages
func GetLang(userLang string, header http.Header) string {
	if userLang != "" {
		return userLang
	}
	langArr := header["Accept-Language"]
	var lang string
	if len(langArr) > 0 {
		lang = langArr[0]
	} else {
		lang = "en"
	}
	return lang
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
		fmt.Println(ind)
		id += string(alphabets[ind])
	}
	return id
}

//ExtractMentions extracts all users mention (@) so that they can be notified
func ExtracMentions(caption string) []string {
	mentions := []string{}
	for i := 0; i < len(caption); i++ {
		if caption[i] == '@' {
			j := i + 1
			for j < len(caption) && caption[j] != ' ' {
				j++
			}
			isValidUserName := validator.IsUserName(caption[i+1 : j+1])
			if isValidUserName {
				mentions = append(mentions, caption[i+1:j+1])
			}
		}
	}
	return mentions
}
