package auth

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"perpus_backend/middleware"
	"perpus_backend/types"
	"perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store types.UserStore
}

func NewHandler(store types.UserStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/login", h.handleLogin).Methods(http.MethodPost)
	r.HandleFunc("/register", h.handleRegister).Methods(http.MethodPost)
	r.HandleFunc("/logout", middleware.AuthWithJWTToken(h.handleLogout, h.store)).Methods(http.MethodPost)
}

// Handler auth login using JWT.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.PayloadLogin

	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadLogin{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	u, err := h.store.GetUserWithRolesByEmail(payload.Email)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("Wrong email!"))
		return
	}

	if !middleware.CompareHashedPassword(u.Password, []byte(payload.Password)) {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("Wrong password!"))
		return
	}

	token, err := middleware.CreateTokenJWT(u.ID, h.store)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":   http.StatusOK,
		"token":  token,
		"status": "OK",
	})
}

// Handle Logout and Revoke the Token using token versioning (user session ver).
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.store.IncrementTokenVersion(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":    http.StatusOK,
		"message": "Successful Logout!",
	})
}

// Handle register user, this not will add the role.
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var (
		payload types.PayloadUser

		fileName string
	)

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadUser{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	file, header, err := r.FormFile("avatar")

	if err == http.ErrMissingFile {
		fileName = "-"
	}

	if err == nil {
		defer file.Close()

		ext := filepath.Ext(header.Filename)
		fileName = utils.Filename + ext

		dst, _ := os.Create("./assets/public/images/" + fileName)
		defer dst.Close()

		io.Copy(dst, file)
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	_, err = h.store.GetUserWithRolesByEmail(payload.Email)
	if err == nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashPass, err := middleware.HashPassword(payload.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.CreateUser(&types.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashPass,
		Avatar:   fileName,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"code":    http.StatusCreated,
		"message": "User Created!",
	})
}
