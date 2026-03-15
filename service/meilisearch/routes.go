package meilisearch

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/perpus_backend/helper"
	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/gorilla/mux"
	"github.com/meilisearch/meilisearch-go"
)

type Handler struct {
	us types.UserStore
	rs types.RoleStore
	ms types.MemberStore
	bs types.BookStore
	cs types.CirculationStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, us types.UserStore, rs types.RoleStore, ms types.MemberStore, bs types.BookStore, cs types.CirculationStore) *Handler {
	return &Handler{
		us:  us,
		rs:  rs,
		ms:  ms,
		bs:  bs,
		cs:  cs,
		jwt: jwt,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/users", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForUsers, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/roles", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForRoles, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/members", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForMembers, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/books", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForBooks, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/circulations", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForCirculations, "admin", "staff", "user"))).Methods(http.MethodGet)
}

func (h *Handler) handleGetSearchForUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	ctx := r.Context()

	query := r.URL.Query().Get("search_u")

	if len(query) < 2 {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("data not found"))
		return
	}

	if !utils.IsIndexMeiliAvailable(ctx, "users") {
		users := h.us.GetUsersForSearch(ctx)

		// assert value users to records meili
		if err := helper.AddDocumentsWithWait("users", "id", users); err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	// initial clientUser meilisearch
	clientUser := utils.NewMSClient

	res, err := clientUser.Index("users").SearchWithContext(ctx, query, &meilisearch.SearchRequest{
		Limit: 10,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("error founded: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JsonResponse{
		Code:   http.StatusOK,
		Data:   res,
		Status: http.StatusText(http.StatusOK),
	})
}

func (h *Handler) handleGetSearchForRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	ctx := r.Context()

	query := r.URL.Query().Get("search_r")

	if len(query) < 2 {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("data not found"))
		return
	}

	if !utils.IsIndexMeiliAvailable(ctx, "roles") {
		roles, _ := h.rs.GetRoles(ctx)

		// assert value roles to records meili
		if err := helper.AddDocumentsWithWait("roles", "id", roles); err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	// initial clientRole meili
	clientRole := utils.NewMSClient

	res, err := clientRole.Index("roles").SearchWithContext(ctx, query, &meilisearch.SearchRequest{
		Limit: 10,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("error founded: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JsonResponse{
		Code:   http.StatusOK,
		Data:   res,
		Status: http.StatusText(http.StatusOK),
	})
}

func (h *Handler) handleGetSearchForMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	ctx := r.Context()

	query := r.URL.Query().Get("search_m")

	if len(query) < 2 {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("data not found"))
		return
	}

	if !utils.IsIndexMeiliAvailable(ctx, "members") {
		members := h.ms.GetMembersForSearch(ctx)

		// assert value members to records meili
		if err := helper.AddDocumentsWithWait("members", "id", members); err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	// initial clientMember meili
	clientMember := utils.NewMSClient

	res, err := clientMember.Index("members").SearchWithContext(ctx, query, &meilisearch.SearchRequest{
		Limit: 10,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("error founded: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JsonResponse{
		Code:   http.StatusOK,
		Data:   res,
		Status: http.StatusText(http.StatusOK),
	})
}

func (h *Handler) handleGetSearchForBooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	ctx := r.Context()

	query := r.URL.Query().Get("search_b")

	if len(query) < 2 {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("data not found"))
		return
	}

	if !utils.IsIndexMeiliAvailable(ctx, "books") {
		books := h.bs.GetBooksForSearch(ctx)

		// assert value books to records meili
		if err := helper.AddDocumentsWithWait("books", "id", books); err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	// initial clientBook meili
	clientBook := utils.NewMSClient

	res, err := clientBook.Index("books").SearchWithContext(ctx, query, &meilisearch.SearchRequest{
		Limit: 10,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("error founded: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JsonResponse{
		Code:   http.StatusOK,
		Data:   res,
		Status: http.StatusText(http.StatusOK),
	})
}

func (h *Handler) handleGetSearchForCirculations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	ctx := r.Context()

	query := r.URL.Query().Get("search_c")

	if len(query) < 2 {
		utils.WriteJSONError(w, http.StatusBadRequest, errors.New("data not found"))
		return
	}

	if !utils.IsIndexMeiliAvailable(ctx, "circulations") {
		circulations := h.cs.GetCirculationsForSearch(ctx)

		// assert value circulations to records meili
		if err := helper.AddDocumentsWithWait("circulations", "id", circulations); err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	// initial clientCirc meili
	clientCirc := utils.NewMSClient

	res, err := clientCirc.Index("circulations").SearchWithContext(ctx, query, &meilisearch.SearchRequest{
		Limit: 10,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("error founded: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.JsonResponse{
		Code:   http.StatusOK,
		Data:   res,
		Status: http.StatusText(http.StatusOK),
	})
}
