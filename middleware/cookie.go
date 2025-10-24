package middleware

import (
	"net/http"
	"perpus_backend/config"
)

// A Cookie represents an HTTP cookie as sent in the Set-Cookie header of an
// HTTP response or the Cookie header of an HTTP request.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("PERPUS")
		if err == http.ErrNoCookie {
			cookie := &http.Cookie{
				Name:     config.Env.CookieName,
				Value:    config.Env.CookieValue,
				Path:     "/",
				Domain:   "localhost",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   3600,
			}

			http.SetCookie(w, cookie)
		}

		next.ServeHTTP(w, r)
	})
}
