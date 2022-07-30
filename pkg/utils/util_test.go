package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetLang(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Add("Accept-Language", "es")
	assert.Equal(t, "en", GetLang("en", c.Request.Header))

	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Add("Accept-Language", "ru")
	assert.Equal(t, "ru", GetLang("", c.Request.Header))

	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Add("Accept-Language", "es")
	assert.Equal(t, "es", GetLang("", c.Request.Header))
}

func TestExtractMentions(t *testing.T) {
	caption := `It's giving dysney land with my Gees @bigjay @hamza come along@brown @somedude 
	@someverylongusername @ls @_shabalala @brown_ @sinp__ing @ adsd`
	expect := []string{"bigjay", "hamza", "somedude"}
	got := ExtracMentions(caption)

	if len(expect) != len(got) {
		t.Fatalf("Testing: %s; want %v got %v", caption, expect, got)
	}

	for i := 0; i < len(expect); i++ {
		if expect[i] != got[i] {
			t.Fatalf("Testing %s; want %s got %s", caption, expect[i], got[i])
		}
	}
}
