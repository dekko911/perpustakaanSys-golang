package user

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/perpus_backend/pkg/hash"
	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	userStore types.UserStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, us types.UserStore) *Handler {
	return &Handler{userStore: us, jwt: jwt}
}

const (
	cok = http.StatusOK

	userProfileFileInPublicPath = "./assets/public/images/profile/"

	size1MB = 1 << 20
)

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/users", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetUsers, "admin"))).Methods(http.MethodGet)

	r.HandleFunc("/users/{userID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetUserWithRolesByID, "admin"))).Methods(http.MethodGet)

	r.HandleFunc("/users", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleCreateUser, "admin"))).Methods(http.MethodPost)

	r.HandleFunc("/users/{userID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleUpdateUser, "admin"))).Methods(http.MethodPut)

	r.HandleFunc("/users/{userID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleDeleteUser, "admin"))).Methods(http.MethodDelete)
}

func (h *Handler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // init context

	page := utils.ParseStringToInt(r.URL.Query().Get("page"))

	users, lastPage, err := h.userStore.GetUsersWithPagination(ctx, page)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:     cok,
		Data:     users,
		LastPage: lastPage,
		Page:     page,
		Status:   http.StatusText(cok),
	})
}

func (h *Handler) handleGetUserWithRolesByID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	ctx := r.Context()

	if err := uuid.Validate(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.userStore.GetUserWithRolesByID(ctx, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Data:   user,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) HandleGetProfileUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := jwt.GetUserIDFromContext(ctx)

	user, err := h.userStore.GetUserWithRolesByID(ctx, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Data:   user,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var filename string

	r.Body = http.MaxBytesReader(w, r.Body, size1MB)

	if err := r.ParseMultipartForm(size1MB); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadUser{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesia(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	// gak perlu pakai redis key email, karena dia dipakai untuk checking aja, bukan di pakai secara di simpan lama gitu
	if _, err := h.userStore.GetUserWithRolesByEmail(ctx, payload.Email); err == nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	file, header, err := r.FormFile("avatar")

	if err == http.ErrMissingFile {
		filename = "-"
	}

	if err == nil {
		defer file.Close()

		filename, err = utils.SetNewFilenameImg("random", file, userProfileFileInPublicPath, header.Filename, header.Size)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	hashPass, err := hash.MakePasswordHash(payload.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	err = h.userStore.CreateUser(ctx, &types.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashPass,
		Avatar:   filename,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonResponse{
		Code:    http.StatusCreated,
		Message: "User Created!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authID := jwt.GetUserIDFromContext(ctx)
	userID := mux.Vars(r)["userID"]

	var filename string

	if r.Method != http.MethodPut {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, errors.New("method doesn't allowed"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, size1MB)

	if err := uuid.Validate(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := r.ParseMultipartForm(size1MB); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadUpdateUser{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesia(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	u, err := h.userStore.GetUserWithRolesByID(ctx, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	for _, r := range u.Roles {
		if r.Name == "admin" {
			if authID != u.ID {
				utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("dilarang edit admin selain admin sendiri"))
				return
			}
		}
	}

	hashPass, err := hash.MakePasswordHash(payload.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
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
		filename = u.Avatar
	}

	if err == nil {
		defer file.Close()

		filename, err = utils.UpdateTheFilenameImg("random", file, userProfileFileInPublicPath, u.Avatar, header.Filename, header.Size)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	err = h.userStore.UpdateUser(ctx, userID, &types.User{
		Name:     u.Name,
		Email:    u.Email,
		Password: u.Password,
		Avatar:   filename,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "User Updated!",
		Status:  http.StatusText(cok),
	})
}

func (h *Handler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	ctx := r.Context()

	if err := uuid.Validate(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	u, err := h.userStore.GetUserWithRolesByID(ctx, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	for _, r := range u.Roles {
		for name := range strings.SplitSeq(r.Name, ",") {
			if name == "admin" {
				utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("you can't delete admin"))
				return
			}
		}
	}

	filePathAndFileName := filepath.Join(userProfileFileInPublicPath, u.Avatar)

	if err := utils.DeleteFilepathWithFilename(filePathAndFileName); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.userStore.DeleteUser(ctx, userID); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "User Deleted!",
		Status:  http.StatusText(cok),
	})
}
