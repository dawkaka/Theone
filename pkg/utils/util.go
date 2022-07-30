package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/dawkaka/theone/pkg/validator"
)

//User's prefered language for success or error messages
func GetLang(userLang string, header http.Header) string {
	if userLang != "" {
		return userLang
	}
	lang := header.Get("Accept-Language")

	if lang != "" {
		return lang
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
		fmt.Println(ind)
		id += string(alphabets[ind])
	}
	return id
}

//ExtractMentions extracts all users mention (@) so that they can be notified
func ExtracMentions(caption string) []string {
	mentions := []string{}
	captionWords := strings.Split(caption, " ")
	for _, val := range captionWords {
		if val[0] == '@' {
			userName := val[1:]
			if validator.IsUserName(userName) {
				mentions = append(mentions, userName)
			}
		}
	}
	return mentions
}
