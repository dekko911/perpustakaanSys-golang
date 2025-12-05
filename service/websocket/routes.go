package websocket

import (
	"net/http"
	"perpus_backend/pkg/jwt"
	"perpus_backend/types"
	"perpus_backend/utils"

	"github.com/gorilla/mux"
	"github.com/meilisearch/meilisearch-go"
)

type Handler struct {
	us types.UserStore
	rs types.RoleStore
	ms types.MemberStore
	bs types.BookStore
	cs types.CirculationStore
}

func NewHandler(us types.UserStore, rs types.RoleStore, ms types.MemberStore, bs types.BookStore, cs types.CirculationStore) *Handler {
	return &Handler{
		us: us,
		rs: rs,
		ms: ms,
		bs: bs,
		cs: cs,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/search/users", jwt.AuthWithJWTToken(jwt.RoleGate(h.us, "admin", "staff", "user")(h.handleGetSearchForUsers), h.us)).Methods(http.MethodGet)

	r.HandleFunc("/search/roles", jwt.AuthWithJWTToken(jwt.RoleGate(h.us, "admin", "staff", "user")(h.handleGetSearchForRoles), h.us)).Methods(http.MethodGet)

	r.HandleFunc("/search/members", jwt.AuthWithJWTToken(jwt.RoleGate(h.us, "admin", "staff", "user")(h.handleGetSearchForMembers), h.us)).Methods(http.MethodGet)

	r.HandleFunc("/search/books", jwt.AuthWithJWTToken(jwt.RoleGate(h.us, "admin", "staff", "user")(h.handleGetSearchForBooks), h.us)).Methods(http.MethodGet)

	r.HandleFunc("/search/circulations", jwt.AuthWithJWTToken(jwt.RoleGate(h.us, "admin", "staff", "user")(h.handleGetSearchForCirculations), h.us)).Methods(http.MethodGet)
}

var req types.SetPayloadQuery

func (h *Handler) handleGetSearchForUsers(w http.ResponseWriter, r *http.Request) {
	conn, err := utils.WSUpgrader.Upgrade(w, r, nil)
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

		// initial index meilisearch
		index := utils.MSClient.Index("users")

		users, _ := h.us.GetUsers()
		if _, err := index.AddDocuments(users, nil); err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		res, err := index.Search(req.QueryUser, &meilisearch.SearchRequest{
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

		// initial index meilisearch
		index := utils.MSClient.Index("roles")

		roles, _ := h.rs.GetRoles()
		if _, err := index.AddDocuments(roles, nil); err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		res, err := index.Search(req.QueryRole, &meilisearch.SearchRequest{
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

		index := utils.MSClient.Index("members")

		members, _ := h.ms.GetMembers()
		if _, err := index.AddDocuments(members, nil); err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		res, err := index.Search(req.QueryMember, &meilisearch.SearchRequest{
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

		index := utils.MSClient.Index("books")

		books, _ := h.bs.GetBooks()
		if _, err := index.AddDocuments(books, nil); err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		res, err := index.Search(req.QueryBook, &meilisearch.SearchRequest{
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

		index := utils.MSClient.Index("circulations")

		circulations, _ := h.cs.GetCirculations()
		if _, err := index.AddDocuments(circulations, nil); err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		res, err := index.Search(req.QueryCirculation, &meilisearch.SearchRequest{
			Limit: 20,
		})
		if err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		conn.WriteJSON(res)
	}
}
