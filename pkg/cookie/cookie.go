package cookie

import (
	"net/http"

	"github.com/perpus_backend/config"
	"github.com/perpus_backend/utils"
)

// A Cookie represents an HTTP cookie as sent in the Set-Cookie header of an
// HTTP response or the Cookie header of an HTTP request.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// search the cookie first
		_, err := r.Cookie(config.Env.CookieName)

		// if cookieKey is not found or not available, create the new one
		if err == http.ErrNoCookie {
			cookie := &http.Cookie{
				Name:     config.Env.CookieName,
				Value:    config.Env.CookieValue,
				Path:     "/",
				Domain:   config.Env.SessionDomain,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   utils.ParseStringToInt(config.Env.CookieMaxAge),
				Secure:   config.Env.AppENV == "production",
			}

			http.SetCookie(w, cookie)
		}

		next.ServeHTTP(w, r)
	})
}
