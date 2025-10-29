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
	"github.com/rs/xid"
)

type Handler struct {
	store types.UserStore
}

const (
	COK = http.StatusOK
	OK  = "OK"
)

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
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("wrong email"))
		return
	}

	if !middleware.CompareHashedPassword(u.Password, []byte(payload.Password)) {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("wrong password"))
		return
	}

	token, err := middleware.CreateTokenJWT(u.ID, h.store)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Status: OK,
		Token:  token,
	})
}

// Handle Logout and Revoke the Token using token versioning (user session ver).
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	if err := h.store.IncrementTokenVersion(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:    COK,
		Message: "You've been Logout!",
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

	file, header, errFile := r.FormFile("avatar")

	if errFile == http.ErrMissingFile {
		fileName = "-"
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	if _, err := h.store.GetUserWithRolesByEmail(payload.Email); err == nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashPass, err := middleware.HashPassword(payload.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if errFile == nil {
		defer file.Close()

		randomString := xid.New().String()

		ext := filepath.Ext(header.Filename)
		fileName = randomString + ext

		dst, _ := os.Create("./assets/public/images/" + fileName)
		defer dst.Close()

		io.Copy(dst, file)
	}

	if err := h.store.CreateUser(&types.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashPass,
		Avatar:   fileName,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonData{
		Code:    http.StatusCreated,
		Message: "User Registered!",
	})
}
