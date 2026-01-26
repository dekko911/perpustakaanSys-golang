package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/perpus_backend/config"
	"github.com/perpus_backend/pkg/hash"
	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store types.UserStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, store types.UserStore) *Handler {
	return &Handler{store: store, jwt: jwt}
}

const cok = http.StatusOK

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/login", h.handleLogin).Methods(http.MethodPost)
	r.HandleFunc("/register", h.handleRegister).Methods(http.MethodPost)
	r.HandleFunc("/logout", h.jwt.AuthWithJWTToken(h.handleLogout)).Methods(http.MethodPost)
}

// Handler auth login using JWT.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1)*time.Second)
	defer cancel()

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	var payload types.SetPayloadJSONLogin

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesian(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	u, err := h.store.GetUserWithRolesByEmail(ctx, payload.Email)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("wrong email"))
		return
	}

	if !hash.CompareHashedPassword(u.Password, payload.Password) {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("wrong password"))
		return
	}

	token, err := h.jwt.CreateTokenJWT(ctx, u.ID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Status: http.StatusText(cok),
		Token:  token,
	})
}

// Handle Logout and Revoke the Token using token version.
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := jwt.GetUserIDFromContext(ctx)

	token := utils.GetTokenFromRequest(r)

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	err := h.store.IncrementTokenVersion(ctx, userID, token)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "You've been Logout!",
		Status:  http.StatusText(cok),
	})
}

// Handle register user, this not will add the role.
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	var payload types.SetPayloadJSONUser

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesian(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	if _, err := h.store.GetUserWithRolesByEmail(ctx, payload.Email); err == nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashPass, err := hash.MakePasswordHash(payload.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.store.CreateUser(ctx, &types.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashPass,
		Avatar:   "-",
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonResponse{
		Code:    http.StatusCreated,
		Message: "User Registered!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) PrivateURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	keyFilepath := mux.Vars(r)["filepath"]

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	clientR2, err := config.R2Storage(ctx)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	output, err := clientR2.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &config.Env.CFBucketName,
		Key:    &keyFilepath,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer output.Body.Close()

	// stream the body
	if _, err := io.Copy(w, output.Body); err == io.ErrUnexpectedEOF {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}
}

func (h *Handler) PublicURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	keyFilepath := mux.Vars(r)["filepath"]

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	clientR2, err := config.R2Storage(ctx)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	output, err := clientR2.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &config.Env.CFBucketName,
		Key:    &keyFilepath,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer output.Body.Close()

	// stream the body
	if _, err := io.Copy(w, output.Body); err == io.ErrUnexpectedEOF {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}
}
