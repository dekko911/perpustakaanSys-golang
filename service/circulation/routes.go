package circulation

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
	store     types.CirculationStore
	userStore types.UserStore
}

const cok = http.StatusOK

func NewHandler(s types.CirculationStore, us types.UserStore) *Handler {
	return &Handler{store: s, userStore: us}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/circulations", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleGetCirculations), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/circulations/{cID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleGetCirculationByID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/circulations", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleCreateCirculation), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/circulations/{cID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleUpdateCirculation), h.userStore)).Methods(http.MethodPatch)

	r.HandleFunc("/circulations/{cID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleDeleteCirculation), h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetCirculations(w http.ResponseWriter, r *http.Request) {
	c, err := h.store.GetCirculations()
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:   cok,
		Data:   c,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleGetCirculationByID(w http.ResponseWriter, r *http.Request) {
	circulationID := mux.Vars(r)["cID"]

	if err := uuid.Validate(circulationID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	c, err := h.store.GetCirculationByID(circulationID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:   cok,
		Data:   c,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleCreateCirculation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	payload := types.SetPayloadCirculation{
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

	if _, err := h.store.GetCirculationByPeminjam(payload.Peminjam); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("peminjam has name: %v been exist", payload.Peminjam))
		return
	}

	err := h.store.CreateCirculation(&types.Circulation{
		BukuID:        payload.BukuID,
		Peminjam:      payload.Peminjam,
		TanggalPinjam: utils.ParseStringToFormatDate(payload.TanggalPinjam),
		JatuhTempo:    utils.ParseStringToFormatDate(payload.JatuhTempo),
		Denda:         utils.ParseStringToFloat(payload.Denda),
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonData{
		Code:    http.StatusCreated,
		Message: "Circulation added!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateCirculation(w http.ResponseWriter, r *http.Request) {
	circulationID := mux.Vars(r)["cID"]

	if err := uuid.Validate(circulationID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	p := types.SetPayloadUpdateCirculation{
		BukuID:        r.FormValue("buku_id"),
		Peminjam:      r.FormValue("peminjam"),
		TanggalPinjam: r.FormValue("tanggal_pinjam"),
		JatuhTempo:    r.FormValue("jatuh_tempo"),
		Denda:         r.FormValue("denda"),
	}

	if err := utils.Validate.Struct(p); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	c, err := h.store.GetCirculationByID(circulationID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if p.BukuID != "" {
		c.BukuID = p.BukuID
	}
	if p.Peminjam != "" {
		c.Peminjam = p.Peminjam
	}
	if p.TanggalPinjam != "" {
		c.TanggalPinjam = utils.ParseStringToFormatDate(p.TanggalPinjam)
	}
	if p.JatuhTempo != "" {
		c.JatuhTempo = utils.ParseStringToFormatDate(p.JatuhTempo)
	}
	if p.Denda != "" {
		c.Denda = utils.ParseStringToFloat(p.Denda)
	}

	err = h.store.UpdateCirculation(circulationID, &types.Circulation{
		BukuID:        c.BukuID,
		Peminjam:      c.Peminjam,
		TanggalPinjam: c.TanggalPinjam,
		JatuhTempo:    c.JatuhTempo,
		Denda:         c.Denda,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:    cok,
		Message: "Circulation updated!",
		Status:  http.StatusText(cok),
	})
}

func (h *Handler) handleDeleteCirculation(w http.ResponseWriter, r *http.Request) {
	circulationID := mux.Vars(r)["cID"]

	if err := uuid.Validate(circulationID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.store.DeleteCirculation(circulationID); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:    cok,
		Message: "Circulation deleted!",
		Status:  http.StatusText(cok),
	})
}
