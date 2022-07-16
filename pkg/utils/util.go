package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

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
