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

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   users,
		Status: OK,
	})
}

func (h *Handler) handleGetUserWithRolesByID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	user, err := h.store.GetUserWithRolesByID(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   user,
		Status: OK,
	})
}

func (h *Handler) handleGetProfileUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	user, err := h.store.GetUserWithRolesByID(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   user,
		Status: OK,
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

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	if _, err := h.store.GetUserWithRolesByEmail(payload.Email); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	file, header, err := r.FormFile("avatar")

	if err == http.ErrMissingFile {
		fileName = "-"
	}

	if err == nil {
		defer file.Close()

		randomString := xid.New().String()

		ext := filepath.Ext(header.Filename)
		fileName = randomString + ext

		dst, _ := os.Create("./assets/public/images/" + fileName)
		defer dst.Close()

		io.Copy(dst, file)
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
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonData{
		Code:    http.StatusCreated,
		Message: "User Created!",
	})
}

func (h *Handler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	var (
		payload types.PayloadForUpdateUser

		fileName string
	)

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, errors.New("method doesn't allowed"))
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

	file, header, err := r.FormFile("avatar")

	if err == http.ErrMissingFile {
		fileName = u.Avatar
	}

	if err == nil {
		defer file.Close()

		fileImg := "./assets/public/images/" + u.Avatar

		if err := os.Remove(fileImg); err != nil {
			utils.WriteJSONError(w, http.StatusNotFound, err)
			return
		}

		randomString := xid.New().String()

		ext := filepath.Ext(header.Filename)
		fileName = randomString + ext

		dst, _ := os.Create("./assets/public/images/" + fileName)
		defer dst.Close()

		io.Copy(dst, file)
	}

	if err := h.store.UpdateUser(userID, &types.User{
		Name:     u.Name,
		Email:    u.Email,
		Password: u.Password,
		Avatar:   fileName,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:    COK,
		Message: "User Updated!",
	})
}

func (h *Handler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

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
		utils.WriteJSONError(w, http.StatusForbidden, errors.New("you can't delete admin"))
		return
	}

	fileName := "./assets/public/images/" + u.Avatar
	os.Remove(fileName)

	if err := h.store.DeleteUser(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:    COK,
		Message: "User Deleted!",
	})
}
