package role

import (
	"net/http"
	"perpus_backend/middleware"
	"perpus_backend/types"
	"perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.RoleStore
	userStore types.UserStore
}

func NewHandler(store types.RoleStore, userStore types.UserStore) *Handler {
	return &Handler{
		store:     store,
		userStore: userStore,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/roles", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleGetRoles), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/roles/{roleID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleGetRoleByID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/roles", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleCreateRole), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/roles/{roleID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleUpdateRole), h.userStore)).Methods(http.MethodPatch)

	r.HandleFunc("/roles/{roleID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleDeleteRole), h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.store.GetRoles()
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":   http.StatusOK,
		"roles":  roles,
		"status": "OK",
	})
}

func (h *Handler) handleGetRoleByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["roleID"]

	role, err := h.store.GetRoleByID(roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":   http.StatusOK,
		"role":   role,
		"status": "OK",
	})
}

func (h *Handler) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	var payload types.PayloadRole

	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadRole{
		Name: r.FormValue("name"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	if err := h.store.CreateRole(&types.Role{
		Name: payload.Name,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"code":    http.StatusCreated,
		"message": "Role Created!",
	})
}

func (h *Handler) handleUpdateRole(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	roleID := vars["roleID"]

	var payload types.PayloadUpdateRole

	if err := req.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadUpdateRole{
		Name: req.FormValue("name"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	r, err := h.store.GetRoleByID(roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if payload.Name != "" {
		r.Name = payload.Name
	}

	if err := h.store.UpdateRole(roleID, &types.Role{
		Name: payload.Name,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":    http.StatusOK,
		"message": "Role Updated!",
	})
}

func (h *Handler) handleDeleteRole(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	roleID := vars["roleID"]

	r, err := h.store.GetRoleByID(roleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	if err := h.store.DeleteRole(r.ID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":    http.StatusOK,
		"message": "Role Deleted!",
	})
}
