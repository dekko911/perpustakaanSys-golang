package circulation

import (
	"net/http"
	"perpus_backend/pkg/jwt"
	"perpus_backend/types"
	"perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.CirculationStore
	userStore types.UserStore
}

const COK = http.StatusOK

func NewHandler(s types.CirculationStore, us types.UserStore) *Handler {
	return &Handler{store: s, userStore: us}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/circulations", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleGetCirculations), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/circulations/{cID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleGetCirculationByID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/circulations", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleCreateCirculation), h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) handleGetCirculations(w http.ResponseWriter, r *http.Request) {
	c, err := h.store.GetCirculations()
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   c,
		Status: http.StatusText(COK),
	})
}

func (h *Handler) handleGetCirculationByID(w http.ResponseWriter, r *http.Request) {
	circulationID := mux.Vars(r)["cID"]

	c, err := h.store.GetCirculationByID(circulationID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   c,
		Status: http.StatusText(COK),
	})
}

func (h *Handler) handleCreateCirculation(w http.ResponseWriter, r *http.Request) {
	var payload types.PayloadCirculation

	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	payload = types.PayloadCirculation{
		BukuID:        r.FormValue("buku_id"),
		Peminjam:      r.FormValue("peminjam"),
		TanggalPinjam: r.FormValue("tanggal_pinjam"),
		JatuhTempo:    r.FormValue("jatuh_tempo"),
		Denda:         r.FormValue("denda"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	// mani lanjut ya man
}
