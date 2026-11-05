package auth

import (
	"net/http"
	"time"
)

// requestIsHTTPS reports whether the client reached us over HTTPS, either
// directly or through a TLS-terminating proxy (Railway, nginx).
func requestIsHTTPS(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

// newAuthCookie builds every auth cookie the same way, so login, refresh,
// logout and password reset can no longer drift apart.
//
// Over HTTPS the cookie is Secure + SameSite=None so it still works when the
// web client calls the API cross-site. Over plain HTTP (local dev) browsers
// reject SameSite=None without Secure, so it falls back to Lax.
//
// A maxAge of 0 or less clears the cookie.
func newAuthCookie(r *http.Request, name, value string, maxAge time.Duration) *http.Cookie {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if maxAge > 0 {
		c.MaxAge = int(maxAge.Seconds())
	} else {
		c.MaxAge = -1
	}
	if requestIsHTTPS(r) {
		c.Secure = true
		c.SameSite = http.SameSiteNoneMode
	}
	return c
}

func setAuthCookie(w http.ResponseWriter, r *http.Request, name, value string, maxAge time.Duration) {
	http.SetCookie(w, newAuthCookie(r, name, value, maxAge))
}

func clearAuthCookie(w http.ResponseWriter, r *http.Request, name string) {
	http.SetCookie(w, newAuthCookie(r, name, "", 0))
}
