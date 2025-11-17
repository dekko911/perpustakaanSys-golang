package roleuser

import (
	"fmt"
	"net/http"
	"perpus_backend/pkg/jwt"
	"perpus_backend/types"
	"perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.RoleUserStore
	userStore types.UserStore
	roleStore types.RoleStore
}

const COK = http.StatusOK

func NewHandler(store types.RoleUserStore, userStore types.UserStore, roleStore types.RoleStore) *Handler {
	return &Handler{store: store, userStore: userStore, roleStore: roleStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/role_user/{userID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin")(h.handleGetUserWithRoleByUserID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/role_user", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin")(h.handleAssignRoleIntoUser), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/user/{userID}/role/{roleID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin")(h.handleDeleteRoleFromUser), h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetUserWithRoleByUserID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	if err := uuid.Validate(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	roleUser, err := h.store.GetUserWithRoleByUserID(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   roleUser,
		Status: http.StatusText(COK),
	})
}

func (h *Handler) handleAssignRoleIntoUser(w http.ResponseWriter, req *http.Request) {
	var payload types.SetPayloadRoleAndUserID

	if err := req.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.SetPayloadRoleAndUserID{
		UserID: req.FormValue("user_id"),
		RoleID: req.FormValue("role_id"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	_, err := h.userStore.GetUserWithRolesByID(payload.UserID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	r, err := h.roleStore.GetRoleByID(payload.RoleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if r.Name == "admin" {
		if jwt.GetUserIDFromContext(req.Context()) != payload.UserID {
			utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("you can't add admin"))
			return
		}
	}

	if err := h.store.AssignRoleIntoUser(payload.UserID, payload.RoleID); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:    COK,
		Message: "User and Role has Connected.",
		Status:  http.StatusText(COK),
	})
}

func (h *Handler) handleDeleteRoleFromUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	roleID := mux.Vars(r)["roleID"]

	// validate id user
	if err := uuid.Validate(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	// validate id role
	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.store.DeleteRoleFromUser(userID, roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JsonData{
		Code:    COK,
		Message: "User and Role has Disconnected.",
		Status:  http.StatusText(COK),
	})
}
