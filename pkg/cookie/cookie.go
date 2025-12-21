package cookie

import (
	"net/http"

	"github.com/perpus_backend/config"
)

// A Cookie represents an HTTP cookie as sent in the Set-Cookie header of an
// HTTP response or the Cookie header of an HTTP request.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie(config.Env.CookieName)
		if err == http.ErrNoCookie {
			// make new cookies if the first cookie isn't available
			cookie := &http.Cookie{
				Name:     config.Env.CookieName,
				Value:    config.Env.CookieValue,
				Path:     "/",
				Domain:   config.Env.SessionDomain,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   3600,
			}

			http.SetCookie(w, cookie)
		}

		next.ServeHTTP(w, r)
	})
}
