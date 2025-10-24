package user

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
	r.HandleFunc("/users", middleware.AuthWithJWTToken(middleware.NeededRole(h.store, "admin")(h.handleGetUsers), h.store)).Methods(http.MethodGet)

	r.HandleFunc("/users/{userID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.store, "admin")(h.handleGetUserWithRolesByID), h.store)).Methods(http.MethodGet)

	r.HandleFunc("/profile", middleware.AuthWithJWTToken(h.handleGetProfileUser, h.store)).Methods(http.MethodGet)

	r.HandleFunc("/users", middleware.AuthWithJWTToken(middleware.NeededRole(h.store, "admin")(h.handleCreateUser), h.store)).Methods(http.MethodPost)

	r.HandleFunc("/users/{userID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.store, "admin")(h.handleUpdateUser), h.store)).Methods(http.MethodPost)

	r.HandleFunc("/users/{userID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.store, "admin")(h.handleDeleteUser), h.store)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.GetUsers()
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":   http.StatusOK,
		"users":  users,
		"status": "OK",
	})
}

func (h *Handler) handleGetUserWithRolesByID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	user, err := h.store.GetUserWithRolesByID(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":   http.StatusOK,
		"user":   user,
		"status": "OK",
	})
}

func (h *Handler) handleGetProfileUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	user, err := h.store.GetUserWithRolesByID(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":   http.StatusOK,
		"user":   user,
		"status": "OK",
	})
}

func (h *Handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
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
		return
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

	if _, err = h.store.GetUserWithRolesByEmail(payload.Email); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashPass, err := middleware.HashPassword(payload.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.store.CreateUser(&types.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashPass,
		Avatar:   fileName,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"code":    http.StatusCreated,
		"message": "User Created!",
	})
}

func (h *Handler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	var (
		payload types.PayloadForUpdateUser

		fileName string
	)

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, errors.New("Method doesn't allowed."))
		return
	}

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadForUpdateUser{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	u, err := h.store.GetUserWithRolesByID(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	file, header, err := r.FormFile("avatar")

	if err == http.ErrMissingFile {
		fileName = u.Avatar
		return
	}

	if err == nil {
		defer file.Close()

		fileImg := "./assets/images/" + fileName

		if err := os.Remove(fileImg); err != nil {
			utils.WriteJSONError(w, http.StatusNotFound, err)
			return
		}

		ext := filepath.Ext(header.Filename)
		fileName = utils.Filename + ext

		dst, _ := os.Create("./assets/public/images/" + fileName)
		defer dst.Close()

		io.Copy(dst, file)
	}

	hashPass, err := middleware.HashPassword(payload.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if payload.Name != "" {
		u.Name = payload.Name
	}
	if payload.Email != "" {
		u.Email = payload.Email
	}
	if payload.Password != "" {
		u.Password = hashPass
	}

	if err := h.store.UpdateUser(userID, &types.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashPass,
		Avatar:   fileName,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":    http.StatusOK,
		"message": "User Updated!",
	})
}

func (h *Handler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	u, err := h.store.GetUserWithRolesByID(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	var role string
	for _, r := range u.Roles {
		role = r.Name
	}

	if role == "admin" {
		utils.WriteJSONError(w, http.StatusForbidden, errors.New("You can't delete admin!"))
		return
	}

	fileName := "./assets/public/images/" + u.Avatar
	if err := os.Remove(fileName); err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	if err := h.store.DeleteUser(u.ID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":    http.StatusOK,
		"message": "User Deleted!",
	})
}
