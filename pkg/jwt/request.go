package jwt

import "net/http"

const AccessTokenCookieName = "access_token"

func TokenFromRequest(r *http.Request) (string, bool) {
	if token, ok := BearerToken(r); ok {
		return token, true
	}

	if token, ok := CookieToken(r, AccessTokenCookieName); ok {
		return token, true
	}

	return "", false
}

func CookieToken(r *http.Request, name string) (string, bool) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", false
	}

	if cookie.Value == "" {
		return "", false
	}

	return cookie.Value, true
}
