package roleuser

import (
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
}

func NewHandler(store types.RoleUserStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/role_user/{userID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleGetRoleByUserID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/role_user", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleAssignRoleIntoUser), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/user/{userID}/role/{roleID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin")(h.handleDeleteRoleFromUser), h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetRoleByUserID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	role, err := h.store.GetRoleByUserID(userID)
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

func (h *Handler) handleAssignRoleIntoUser(w http.ResponseWriter, r *http.Request) {
	var payload types.PayloadRoleUserID

	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadRoleUserID{
		UserID: r.FormValue("user_id"),
		RoleID: r.FormValue("role_id"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	if err := h.store.AssignRoleIntoUser(payload.UserID, payload.RoleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"code":    http.StatusCreated,
		"message": "User And Role has Connected.",
	})
}

func (h *Handler) handleDeleteRoleFromUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	roleID := mux.Vars(r)["roleID"]

	if err := h.store.DeleteRoleFromUser(userID, roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"code":    http.StatusOK,
		"message": "User And Role Has Disconnected.",
	})
}
