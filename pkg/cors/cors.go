package cors

import (
	"fmt"
	"net/http"
	"perpus_backend/config"
)

// Allows web pages to securely access resources from other domains, overcoming the same-origin policy restrictions that apply by default.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("%s:%s", config.Env.AppURL, "5173"))
		w.Header().Set("Vary", "Origin")

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method != http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
