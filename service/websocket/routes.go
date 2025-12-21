package websocket

import (
	"context"
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
	r.HandleFunc("/search/users", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForUsers, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/search/roles", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForRoles, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/search/members", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForMembers, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/search/books", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForBooks, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/search/circulations", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetSearchForCirculations, "admin", "staff", "user"))).Methods(http.MethodGet)
}

var req types.SetPayloadQuery

func (h *Handler) handleGetSearchForUsers(w http.ResponseWriter, r *http.Request) {
	conn, err := utils.WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	// initial clientUser meilisearch
	clientUser := utils.MSClient

	// assert value users to records meili
	users, _ := h.us.GetUsers(context.Background())

	err = helper.AddDocumentsWithWait(clientUser, "users", "id", users)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer conn.Close()

	for {
		if err := conn.ReadJSON(&req); err != nil {
			conn.WriteJSON("error read payload json")
			return
		}

		if len(req.QueryUser) < 1 {
			conn.WriteJSON("data not found")
			continue
		}

		res, err := clientUser.Index("users").Search(req.QueryUser, &meilisearch.SearchRequest{
			Limit: 20,
		})
		if err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		conn.WriteJSON(res)
	}
}

func (h *Handler) handleGetSearchForRoles(w http.ResponseWriter, r *http.Request) {
	conn, err := utils.WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	// initial clientRole meili
	clientRole := utils.MSClient

	// assert value roles to records meili
	roles, _ := h.rs.GetRoles(context.Background())

	err = helper.AddDocumentsWithWait(clientRole, "roles", "id", roles)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer conn.Close()

	for {
		if err := conn.ReadJSON(&req); err != nil {
			conn.WriteJSON("error read payload json")
			return
		}

		if len(req.QueryRole) < 1 {
			conn.WriteJSON("data not found")
			continue
		}

		res, err := clientRole.Index("roles").Search(req.QueryRole, &meilisearch.SearchRequest{
			Limit: 20,
		})
		if err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		conn.WriteJSON(res)
	}
}

func (h *Handler) handleGetSearchForMembers(w http.ResponseWriter, r *http.Request) {
	conn, err := utils.WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	// initial clientMember meili
	clientMember := utils.MSClient

	// assert value members to records meili
	members, _ := h.ms.GetMembers(context.Background())

	err = helper.AddDocumentsWithWait(clientMember, "members", "id", members)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer conn.Close()

	for {
		if err := conn.ReadJSON(&req); err != nil {
			conn.WriteJSON("error read payload json")
			return
		}

		if len(req.QueryMember) < 1 {
			conn.WriteJSON("data not found")
			continue
		}

		res, err := clientMember.Index("members").Search(req.QueryMember, &meilisearch.SearchRequest{
			Limit: 20,
		})
		if err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		conn.WriteJSON(res)
	}
}

func (h *Handler) handleGetSearchForBooks(w http.ResponseWriter, r *http.Request) {
	conn, err := utils.WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	// initial clientBook meili
	clientBook := utils.MSClient

	// assert value books to records meili
	books, _ := h.bs.GetBooks(context.Background())

	err = helper.AddDocumentsWithWait(clientBook, "books", "id", books)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer conn.Close()

	for {
		if err := conn.ReadJSON(&req); err != nil {
			conn.WriteJSON("error read payload json")
			return
		}

		if len(req.QueryBook) < 1 {
			conn.WriteJSON("data not found")
			continue
		}

		res, err := clientBook.Index("books").Search(req.QueryBook, &meilisearch.SearchRequest{
			Limit: 20,
		})
		if err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		conn.WriteJSON(res)
	}
}

func (h *Handler) handleGetSearchForCirculations(w http.ResponseWriter, r *http.Request) {
	conn, err := utils.WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	// initial clientCirc meili
	clientCirc := utils.MSClient

	// assert value circulations to records meili
	circulations, _ := h.cs.GetCirculations(context.Background())

	err = helper.AddDocumentsWithWait(clientCirc, "circulations", "id", circulations)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer conn.Close()

	for {
		if err := conn.ReadJSON(&req); err != nil {
			conn.WriteJSON("error read payload json")
			return
		}

		if len(req.QueryCirculation) < 1 {
			conn.WriteJSON("data not found")
			continue
		}

		res, err := clientCirc.Index("circulations").Search(req.QueryCirculation, &meilisearch.SearchRequest{
			Limit: 20,
		})
		if err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		conn.WriteJSON(res)
	}
}
