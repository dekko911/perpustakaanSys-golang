package user

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"perpus_backend/pkg/hash"
	"perpus_backend/pkg/jwt"
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

	filePublicPath = "./assets/public/images/profile/"

	size1MB = 1 << 20
)

func NewHandler(store types.UserStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/users", jwt.AuthWithJWTToken(jwt.NeededRole(h.store, "admin")(h.handleGetUsers), h.store)).Methods(http.MethodGet)

	r.HandleFunc("/users/{userID}", jwt.AuthWithJWTToken(jwt.NeededRole(h.store, "admin")(h.handleGetUserWithRolesByID), h.store)).Methods(http.MethodGet)

	r.HandleFunc("/profile", jwt.AuthWithJWTToken(h.handleGetProfileUser, h.store)).Methods(http.MethodGet)

	r.HandleFunc("/users", jwt.AuthWithJWTToken(jwt.NeededRole(h.store, "admin")(h.handleCreateUser), h.store)).Methods(http.MethodPost)

	r.HandleFunc("/users/{userID}", jwt.AuthWithJWTToken(jwt.NeededRole(h.store, "admin")(h.handleUpdateUser), h.store)).Methods(http.MethodPut)

	r.HandleFunc("/users/{userID}", jwt.AuthWithJWTToken(jwt.NeededRole(h.store, "admin")(h.handleDeleteUser), h.store)).Methods(http.MethodDelete)
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
		Status: http.StatusText(COK),
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
		Status: http.StatusText(COK),
	})
}

func (h *Handler) handleGetProfileUser(w http.ResponseWriter, r *http.Request) {
	userID := jwt.GetUserIDFromContext(r.Context())

	user, err := h.store.GetUserWithRolesByID(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   user,
		Status: http.StatusText(COK),
	})
}

func (h *Handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var (
		payload types.PayloadUser

		fileName, filePath string
		sizeFile           int64
	)

	if err := r.ParseMultipartForm(size1MB); err != nil {
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
		sizeFile = header.Size

		if sizeFile <= size1MB {
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				fileName = randomString + ext
				filePath = filePublicPath + fileName

				dst, _ := os.Create(filePath)
				defer dst.Close()

				io.Copy(dst, file)
			} else {
				utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("only support jpg, jpeg, and png"))
				return
			}
		} else {
			utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("only serve file under 1mb"))
			return
		}
	}

	hashPass, err := hash.HashPassword(payload.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.store.CreateUser(&types.User{
		Name:     payload.Name,  // miko
		Email:    payload.Email, // miko@email.com
		Password: hashPass,      // miko123
		Avatar:   fileName,      // img.jpg
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonData{
		Code:    http.StatusCreated,
		Message: "User Created!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	var (
		payload types.PayloadUpdateUser

		fileName, filePath string
		sizeFile           int64
	)

	if r.Method != http.MethodPut {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, errors.New("method doesn't allowed"))
		return
	}

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadUpdateUser{
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

	hashPass, err := hash.HashPassword(payload.Password)
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

		randomString := xid.New().String()

		ext := filepath.Ext(header.Filename)
		sizeFile = header.Size

		if sizeFile <= size1MB {
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				fileImgOld := filePublicPath + u.Avatar

				if err := os.Remove(fileImgOld); err != nil {
					utils.WriteJSONError(w, http.StatusNotFound, err)
					return
				}

				fileName = randomString + ext
				filePath = filePublicPath + fileName

				dst, _ := os.Create(filePath)
				defer dst.Close()

				io.Copy(dst, file)
			} else {
				utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("only support jpg, jpeg, and png"))
				return
			}
		} else {
			utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("only serve file under 1mb"))
			return
		}
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
		Status:  http.StatusText(COK),
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

	fileName := filePublicPath + u.Avatar
	info, err := os.Stat(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			utils.WriteJSONError(w, http.StatusNotFound, fmt.Errorf("file not found"))
			return
		}

		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}
	if !info.IsDir() {
		os.Remove(fileName)
	}

	if err := h.store.DeleteUser(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:    COK,
		Message: "User Deleted!",
		Status:  http.StatusText(COK),
	})
}
