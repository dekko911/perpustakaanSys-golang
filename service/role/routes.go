package role

import (
	"errors"
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
	roleStore types.RoleStore
	userStore types.UserStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, rs types.RoleStore, us types.UserStore) *Handler {
	return &Handler{
		roleStore: rs,
		userStore: us,
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

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	roles, err := h.roleStore.GetRoles(ctx)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Data:   roles,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleGetRoleByID(w http.ResponseWriter, r *http.Request) {
	roleID := mux.Vars(r)["roleID"]

	ctx := r.Context()

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	role, err := h.roleStore.GetRoleByID(ctx, roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Data:   role,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadRole{
		Name: r.PostForm.Get("name"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesia(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	// if the role name was out of the box, it should be triggered
	if !utils.IsInputRoleNameWasValid(payload.Name) {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("invalid role name; only admin, staff, and user can be valid"))
		return
	}

	if _, err := h.roleStore.GetRoleByName(ctx, payload.Name); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("role with name: %s is already exists", payload.Name))
		return
	}

	err := h.roleStore.CreateRole(ctx, &types.Role{
		Name: payload.Name,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonResponse{
		Code:    http.StatusCreated,
		Message: "Role Created!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateRole(w http.ResponseWriter, req *http.Request) {
	roleID := mux.Vars(req)["roleID"]

	ctx := req.Context()

	if req.Method != http.MethodPatch {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", req.Method))
		return
	}

	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := req.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadUpdateRole{
		Name: req.PostForm.Get("name"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesia(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	r, err := h.roleStore.GetRoleByID(ctx, roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	if payload.Name != "" {
		r.Name = payload.Name
	}

	if !utils.IsInputRoleNameWasValid(r.Name) {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("invalid role name; only admin, staff, and user can be valid"))
		return
	}

	err = h.roleStore.UpdateRole(ctx, roleID, &types.Role{
		Name: r.Name,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "Role Updated!",
		Status:  http.StatusText(cok),
	})
}

func (h *Handler) handleDeleteRole(w http.ResponseWriter, req *http.Request) {
	roleID := mux.Vars(req)["roleID"]

	ctx := req.Context()

	if req.Method != http.MethodDelete {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", req.Method))
		return
	}

	if err := uuid.Validate(roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	r, err := h.roleStore.GetRoleByID(ctx, roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if r.Name == "admin" {
		utils.WriteJSONError(w, http.StatusForbidden, errors.New("you can't delete role admin"))
		return
	}

	if err := h.roleStore.DeleteRole(ctx, roleID); err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "Role Deleted!",
		Status:  http.StatusText(cok),
	})
}
