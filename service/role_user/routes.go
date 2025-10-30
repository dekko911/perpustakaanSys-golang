package roleuser

import (
	"fmt"
	"net/http"
	"perpus_backend/middleware"
	"perpus_backend/types"
	"perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.RoleUserStore
	userStore types.UserStore
	roleStore types.RoleStore
}

const (
	COK = http.StatusOK
	OK  = "OK"
)

func NewHandler(store types.RoleUserStore, userStore types.UserStore, roleStore types.RoleStore) *Handler {
	return &Handler{store: store, userStore: userStore, roleStore: roleStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/role_user/{userID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleGetRoleByUserID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/role_user", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleAssignRoleIntoUser), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/user/{userID}/role/{roleID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleDeleteRoleFromUser), h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetRoleByUserID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	roleUser, err := h.store.GetUserWithRoleByUserID(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   roleUser,
		Status: OK,
	})
}

func (h *Handler) handleAssignRoleIntoUser(w http.ResponseWriter, req *http.Request) {
	var payload types.PayloadRoleUserID

	if err := req.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadRoleUserID{
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
		utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("you can't add admin"))
		return
	}

	if err := h.store.AssignRoleIntoUser(payload.UserID, payload.RoleID); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:    COK,
		Message: "User and Role has Connected.",
	})
}

func (h *Handler) handleDeleteRoleFromUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	roleID := mux.Vars(r)["roleID"]

	if err := h.store.DeleteRoleFromUser(userID, roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JsonData{
		Code:    COK,
		Message: "User and Role has Disconnected.",
	})
}
