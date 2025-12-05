package role

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
	store     types.RoleStore
	userStore types.UserStore
}

const cok = http.StatusOK

func NewHandler(store types.RoleStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/roles", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin")(h.handleGetRoles), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/roles/{roleID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin")(h.handleGetRoleByID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/roles", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin")(h.handleCreateRole), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/roles/{roleID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin")(h.handleUpdateRole), h.userStore)).Methods(http.MethodPatch)

	r.HandleFunc("/roles/{roleID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin")(h.handleDeleteRole), h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.store.GetRoles()
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:   cok,
		Data:   roles,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleGetRoleByID(w http.ResponseWriter, r *http.Request) {
	roleID := mux.Vars(r)["roleID"]

	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	role, err := h.store.GetRoleByID(roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:   cok,
		Data:   role,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadRole{
		Name: r.FormValue("name"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	if _, err := h.store.GetRoleByName(payload.Name); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("role with name: %s is already exists", payload.Name))
		return
	}

	// if the role name was out of the box, it should be triggered
	if !utils.IsInputRoleNameWasValid(payload.Name) {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid role name; only admin, staff, and user can be valid"))
		return
	}

	if err := h.store.CreateRole(types.Role{
		Name: payload.Name,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonData{
		Code:    http.StatusCreated,
		Message: "Role Created!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateRole(w http.ResponseWriter, req *http.Request) {
	roleID := mux.Vars(req)["roleID"]

	var payload types.SetPayloadUpdateRole

	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := req.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.SetPayloadUpdateRole{
		Name: req.FormValue("name"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	r, err := h.store.GetRoleByID(roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	if payload.Name != "" {
		r.Name = payload.Name
	}

	if !utils.IsInputRoleNameWasValid(r.Name) {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid role name; only admin, staff, and user can be valid"))
		return
	}

	if err := h.store.UpdateRole(roleID, types.Role{
		Name: r.Name,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:    cok,
		Message: "Role Updated!",
		Status:  http.StatusText(cok),
	})
}

func (h *Handler) handleDeleteRole(w http.ResponseWriter, req *http.Request) {
	roleID := mux.Vars(req)["roleID"]

	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	r, err := h.store.GetRoleByID(roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if r.Name == "admin" {
		utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("you can't delete role admin"))
		return
	}

	if err := h.store.DeleteRole(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:    cok,
		Message: "Role Deleted!",
		Status:  http.StatusText(cok),
	})
}
