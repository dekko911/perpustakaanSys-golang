package circulation

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
	circulationStore types.CirculationStore
	userStore        types.UserStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, cs types.CirculationStore, us types.UserStore) *Handler {
	return &Handler{
		circulationStore: cs,
		userStore:        us,
		jwt:              jwt,
	}
}

const cok = http.StatusOK

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/circulations", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetCirculations, "admin", "staff"))).Methods(http.MethodGet)

	r.HandleFunc("/circulations/{cID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetCirculationByID, "admin", "staff"))).Methods(http.MethodGet)

	r.HandleFunc("/circulations", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleCreateCirculation, "admin", "staff"))).Methods(http.MethodPost)

	r.HandleFunc("/circulations/{cID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleUpdateCirculation, "admin", "staff"))).Methods(http.MethodPatch)

	r.HandleFunc("/circulations/{cID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleDeleteCirculation, "admin", "staff"))).Methods(http.MethodDelete)
}

func (h *Handler) handleGetCirculations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := utils.ParseStringToInt(r.URL.Query().Get("page"))

	circulations, lastPage, err := h.circulationStore.GetCirculationsWithPagination(ctx, page)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:     cok,
		Data:     circulations,
		Page:     page,
		LastPage: lastPage,
		Status:   http.StatusText(cok),
	})
}

func (h *Handler) handleGetCirculationByID(w http.ResponseWriter, r *http.Request) {
	circulationID := mux.Vars(r)["cID"]

	ctx := r.Context()

	if err := uuid.Validate(circulationID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	circulation, err := h.circulationStore.GetCirculationByID(ctx, circulationID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Data:   circulation,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleCreateCirculation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	if err := uuid.Validate(payload.BukuID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesia(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	if _, err := h.circulationStore.GetCirculationByPeminjam(ctx, payload.Peminjam); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("peminjam has name: %v been exist", payload.Peminjam))
		return
	}

	err := h.circulationStore.CreateCirculation(ctx, &types.Circulation{
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

	utils.WriteJSON(w, http.StatusCreated, utils.JsonResponse{
		Code:    http.StatusCreated,
		Message: "Circulation added!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateCirculation(w http.ResponseWriter, r *http.Request) {
	circulationID := mux.Vars(r)["cID"]

	ctx := r.Context()

	if err := uuid.Validate(circulationID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadUpdateCirculation{
		BukuID:        r.FormValue("buku_id"),
		Peminjam:      r.FormValue("peminjam"),
		TanggalPinjam: r.FormValue("tanggal_pinjam"),
		JatuhTempo:    r.FormValue("jatuh_tempo"),
		Denda:         r.FormValue("denda"),
	}

	if err := uuid.Validate(payload.BukuID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesia(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	c, err := h.circulationStore.GetCirculationByID(ctx, circulationID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if payload.BukuID != "" {
		c.BukuID = payload.BukuID
	}
	if payload.Peminjam != "" {
		c.Peminjam = payload.Peminjam
	}
	if payload.TanggalPinjam != "" {
		c.TanggalPinjam = utils.ParseStringToFormatDate(payload.TanggalPinjam)
	}
	if payload.JatuhTempo != "" {
		c.JatuhTempo = utils.ParseStringToFormatDate(payload.JatuhTempo)
	}
	if payload.Denda != "" {
		c.Denda = utils.ParseStringToFloat(payload.Denda)
	}

	err = h.circulationStore.UpdateCirculation(ctx, circulationID, &types.Circulation{
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

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "Circulation updated!",
		Status:  http.StatusText(cok),
	})
}

func (h *Handler) handleDeleteCirculation(w http.ResponseWriter, r *http.Request) {
	circulationID := mux.Vars(r)["cID"]

	ctx := r.Context()

	if err := uuid.Validate(circulationID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.circulationStore.DeleteCirculation(ctx, circulationID); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "Circulation deleted!",
		Status:  http.StatusText(cok),
	})
}
