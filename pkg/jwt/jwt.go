package jwt

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/perpus_backend/config"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type AuthJWT struct {
	us types.UserStore

	rdb *redis.Client
}

func NewAuthJWT(us types.UserStore, rdb *redis.Client) *AuthJWT {
	return &AuthJWT{us: us, rdb: rdb}
}

type contextKey string // 16 byte string

const (
	userKey contextKey = "userID"

	unauth int = http.StatusUnauthorized
)

var (
	shortTimeoutDuration = time.Duration(3 * time.Second)

	ua = errors.New("Unauthorized")
)

// Authentication using JWT.
func (j *AuthJWT) AuthWithJWTToken(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// initial context for http.request start from here
		ctx := r.Context()

		tokenString := utils.GetTokenFromRequest(r)

		token, err := j.validateTokenJWT(tokenString)
		if err != nil {
			utils.WriteJSONError(w, unauth, ua)
			log.Println(err)
			return
		}

		if !token.Valid {
			utils.WriteJSONError(w, unauth, ua)
			log.Println("token is invalid")
			return
		}

		// get the key in redis storage
		resInt64, err := j.rdb.Exists(ctx, tokenString).Result()
		if err != nil {
			utils.WriteJSONError(w, unauth, err)
			log.Println(err)
			return
		}

		// check once again to make sure the token is available in redis storage
		if resInt64 == 0 {
			utils.WriteJSONError(w, unauth, ua)
			log.Println("token is deleted in redis storage")
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID, ok := claims["userID"].(string)
		if !ok {
			utils.WriteJSONError(w, unauth, ua)
			log.Println("type assertion failed")
			return
		}

		ctx, cancel := context.WithTimeout(ctx, shortTimeoutDuration)
		defer cancel()

		u, err := j.us.GetUserWithRolesByID(ctx, userID)
		if err != nil {
			utils.WriteJSONError(w, unauth, ua)
			log.Println(err)
			return
		}

		if u.TokenVersion != int(claims["token_version"].(float64)) {
			utils.WriteJSONError(w, unauth, ua)
			log.Println("token is revoked")
			return
		}

		if time.Now().Unix() > int64(claims["exp"].(float64)) {
			utils.WriteJSONError(w, unauth, ua)
			log.Println("token is expired")
			return
		}

		ctx = context.WithValue(ctx, userKey, u.ID) // set the value with key userID

		h(w, r.WithContext(ctx)) // <- di mana ada WithContext(), di sana parent context nya.
	}
}

// Creating token for use in method AuthJWT and add at header authentication.
func (j *AuthJWT) CreateTokenJWT(ctx context.Context, userID string) (string, error) {
	u, err := j.us.GetUserWithRolesByID(ctx, userID)
	if err != nil {
		return "", err
	}

	var roles string // initial first
	for _, role := range u.Roles {
		roles = role.Name // get the all role from user
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

	// set token and save into redis storage
	_ = j.rdb.SetEx(ctx, tokenString, u.Name, time.Duration(4*time.Hour)).Err()

	return tokenString, nil
}

func (j *AuthJWT) validateTokenJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.Env.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
}

// get user login info from ctx.
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(userKey).(string); ok {
		return userID
	}

	return ""
}

// using for blocking routes who doesn't have any roles.
// make sure in params roles input values role at there.
func (j *AuthJWT) RoleGate(h http.HandlerFunc, roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := GetUserIDFromContext(ctx)
		if userID == "" {
			utils.WriteJSONError(w, unauth, ua)
			return
		}

		u, err := j.us.GetUserWithRolesByID(ctx, userID)
		if err != nil {
			utils.WriteJSONError(w, unauth, ua)
			return
		}

		for _, role := range u.Roles {
			targetRoles := strings.Split(role.Name, ", ")

			if utils.CompareRole(roles, targetRoles) {
				h(w, r)
				return
			}
		}

		utils.WriteJSONError(w, unauth, ua)
	}
}
