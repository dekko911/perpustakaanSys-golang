package roleuser

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	roleUserStore types.RoleUserStore
	userStore     types.UserStore
	roleStore     types.RoleStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, rus types.RoleUserStore, us types.UserStore, rs types.RoleStore) *Handler {
	return &Handler{
		roleUserStore: rus,
		userStore:     us,
		roleStore:     rs,
		jwt:           jwt,
	}
}

const cok = http.StatusOK

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/role_user/{userID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetUserWithRoleByUserID, "admin"))).Methods(http.MethodGet)

	r.HandleFunc("/role/user/{userID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetUserAndRoleNames, "admin"))).Methods(http.MethodGet)

	r.HandleFunc("/role_user", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleAssignRoleIntoUser, "admin"))).Methods(http.MethodPost)

	r.HandleFunc("/user/{userID}/role/{roleID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleDeleteRoleFromUser, "admin"))).Methods(http.MethodDelete)
}

func (h *Handler) handleGetUserWithRoleByUserID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	ctx := r.Context()

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := uuid.Validate(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	roleUser, err := h.roleUserStore.GetUserWithRoleByUserID(ctx, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Data:   roleUser,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleGetUserAndRoleNames(w http.ResponseWriter, req *http.Request) {
	userID := mux.Vars(req)["userID"]

	ctx := req.Context()

	if req.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", req.Method))
		return
	}

	if err := uuid.Validate(userID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	u, r, err := h.roleUserStore.GetUserAndRoleNames(ctx, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	userRoleMap := map[string]any{
		"user": u,
		"role": r,
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Data:   userRoleMap,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleAssignRoleIntoUser(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", req.Method))
		return
	}

	if err := req.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadRoleAndUserID{
		UserID: req.PostForm.Get("user_id"),
		RoleID: req.PostForm.Get("role_id"),
	}

	// validate id user
	if err := uuid.Validate(payload.UserID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	// validate id role
	if err := uuid.Validate(payload.RoleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesia(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	_, err := h.userStore.GetUserWithRolesByID(ctx, payload.UserID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	r, err := h.roleStore.GetRoleByID(ctx, payload.RoleID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	if r.Name == "admin" {
		if jwt.GetUserIDFromContext(ctx) != payload.UserID {
			utils.WriteJSONError(w, http.StatusForbidden, errors.New("you can't add admin"))
			return
		}
	}

	if err := h.roleUserStore.AssignRoleIntoUser(ctx, payload.UserID, payload.RoleID); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "User and Role has Connected.",
		Status:  http.StatusText(cok),
	})
}

func (h *Handler) handleDeleteRoleFromUser(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	userID := mux.Vars(req)["userID"]
	roleID := mux.Vars(req)["roleID"]

	if req.Method != http.MethodDelete {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", req.Method))
		return
	}

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

	u, err := h.userStore.GetUserWithRolesByID(ctx, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	for _, r := range u.Roles {
		for name := range strings.SplitSeq(r.Name, ",") {
			if name == "admin" {
				utils.WriteJSONError(w, http.StatusForbidden, errors.New("you can't delete admin"))
				return
			}
		}
	}

	if err := h.roleUserStore.DeleteRoleFromUser(ctx, userID, roleID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JsonResponse{
		Code:    cok,
		Message: "User and Role has Disconnected.",
		Status:  http.StatusText(cok),
	})
}
