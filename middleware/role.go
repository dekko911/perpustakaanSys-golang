package middleware

import (
	"log"
	"net/http"
	"perpus_backend/types"
	"slices"
)

// using for blocking routes who doesn't have any roles.
// make sure in params roles, input values role at there.
func NeededRole(us types.UserStore, roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserIDFromContext(r.Context())
			if userID == "" {
				log.Println("userID is empty.")
				permissionDenied(w)
				return
			}

			user, err := us.GetUserWithRolesByID(userID)
			if err != nil {
				log.Printf("user not found, error: %v", err)
				permissionDenied(w)
				return
			}

			for _, role := range user.Roles {
				// check if role.Name has same value with roles param.
				if slices.Contains(roles, role.Name) {
					hf(w, r)
					return
				}
			}

			log.Printf("user %s doesn't have role.", user.Name)
			permissionDenied(w)
		}
	}
}
