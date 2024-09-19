package rest

import (
	"hcm/pkg/criteria/constant"
	"net/url"

	"github.com/emicklei/go-restful/v3"
)

// GetLanguageByHTTPRequest ...
func GetLanguageByHTTPRequest(req *restful.Request) constant.Language {
	cookie, err := req.Request.Cookie(constant.BKHTTPCookieLanguageKey)
	if err != nil {
		return constant.Chinese
	}
	cookieLanguage, _ := url.QueryUnescape(cookie.Value)
	if cookieLanguage == "" {
		return constant.Chinese
	}

	return constant.Language(cookieLanguage)
}
