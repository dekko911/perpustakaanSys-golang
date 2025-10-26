package middleware

import (
	"errors"
	"net/http"
	"perpus_backend/utils"

	"golang.org/x/time/rate"
)

// Limiter request get into routes. Param timer for set when request has been limit.
// Param attempts for how many attempts you want to. ex: 10, 20, 66, and etc.
func RateLimitMiddleware(timer rate.Limit, attempts int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(timer, attempts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				utils.WriteJSONError(w, http.StatusTooManyRequests, errors.New("too many requests"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
