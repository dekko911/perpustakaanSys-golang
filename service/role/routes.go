package role

import (
	"fmt"
	"net/http"

	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.RoleStore
	userStore types.UserStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, store types.RoleStore, userStore types.UserStore) *Handler {
	return &Handler{
		store:     store,
		userStore: userStore,
		jwt:       jwt,
	}
}

const cok = http.StatusOK

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/roles", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetRoles, "admin"))).Methods(http.MethodGet)

	r.HandleFunc("/roles/{roleID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetRoleByID, "admin"))).Methods(http.MethodGet)

	r.HandleFunc("/roles", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleCreateRole, "admin"))).Methods(http.MethodPost)

	r.HandleFunc("/roles/{roleID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleUpdateRole, "admin"))).Methods(http.MethodPatch)

	r.HandleFunc("/roles/{roleID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleDeleteRole, "admin"))).Methods(http.MethodDelete)
}

func (h *Handler) handleGetRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roles, err := h.store.GetRoles(ctx)
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

	ctx := r.Context()

	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	role, err := h.store.GetRoleByID(ctx, roleID)
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
	ctx := r.Context()

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

	if _, err := h.store.GetRoleByName(ctx, payload.Name); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("role with name: %s is already exists", payload.Name))
		return
	}

	// if the role name was out of the box, it should be triggered
	if !utils.IsInputRoleNameWasValid(payload.Name) {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid role name; only admin, staff, and user can be valid"))
		return
	}

	err := h.store.CreateRole(ctx, &types.Role{
		Name: payload.Name,
	})
	if err != nil {
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

	ctx := req.Context()

	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := req.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadUpdateRole{
		Name: req.FormValue("name"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	r, err := h.store.GetRoleByID(ctx, roleID)
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

	err = h.store.UpdateRole(ctx, roleID, &types.Role{
		Name: r.Name,
	})
	if err != nil {
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

	ctx := req.Context()

	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	r, err := h.store.GetRoleByID(ctx, roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if r.Name == "admin" {
		utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("you can't delete role admin"))
		return
	}

	if err := h.store.DeleteRole(ctx, roleID); err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:    cok,
		Message: "Role Deleted!",
		Status:  http.StatusText(cok),
	})
}
