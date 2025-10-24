package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"perpus_backend/config"
	"perpus_backend/types"
	"perpus_backend/utils"
	"slices"
	"time"

	"github.com/golang-jwt/jwt"
)

type contextKey string

const UserKey contextKey = "userID"

// Authentication using JWT.
func AuthWithJWTToken(h http.HandlerFunc, us types.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := utils.GetTokenFromRequest(r)

		token, err := validateTokenJWT(tokenString)
		if err != nil {
			log.Printf("failed to validate token: %v", err)
			permissionDenied(w)
			return
		}

		if !token.Valid {
			log.Println("invalid token.")
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID, ok := claims["userID"].(string)
		if !ok {
			log.Println("failed type assert userID.")
			permissionDenied(w)
			return
		}

		u, err := us.GetUserWithRolesByID(userID)
		if err != nil {
			log.Printf("user not found, error: %v", err)
			permissionDenied(w)
			return
		}

		if u.TokenVersion != int(claims["token_version"].(float64)) {
			log.Println("token has revoked.")
			permissionDenied(w)
			return
		}

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			log.Println("token has been expired.")
			permissionDenied(w)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, u.ID)

		h(w, r.WithContext(ctx))
	}
}

// Creating token for use in method AuthJWT and add at header authentication.
func CreateTokenJWT(userID string, us types.UserStore) (string, error) {
	u, err := us.GetUserWithRolesByID(userID)
	if err != nil {
		return "", err
	}

	var roles string
	for _, role := range u.Roles {
		roles = role.Name
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":        u.ID,
		"roles":         roles,
		"token_version": u.TokenVersion,
		"iat":           time.Now().Unix(),
		"exp":           time.Now().Add(4 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.Env.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateTokenJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(config.Env.JWTSecret), nil
	})
}

func permissionDenied(w http.ResponseWriter) {
	utils.WriteJSONError(w, http.StatusForbidden, errors.New("permission denied."))
}

func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(UserKey).(string)
	if !ok {
		return ""
	}

	return userID
}

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

			log.Printf("user %s doesn't have required role.", user.Name)
			permissionDenied(w)
		}
	}
}
