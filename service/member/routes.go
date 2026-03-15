package member

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
	memberStore types.MemberStore
	userStore   types.UserStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, ms types.MemberStore, us types.UserStore) *Handler {
	return &Handler{
		memberStore: ms,
		userStore:   us,
		jwt:         jwt,
	}
}

const (
	cok = http.StatusOK // for alias http.StatusOK

	r2MembersAvatarPath = "members/avatar"

	size1MB = 1 << 20
)

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/members", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetMembers, "admin", "staff"))).Methods(http.MethodGet)

	r.HandleFunc("/member/{memberID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetMemberByID, "admin", "staff"))).Methods(http.MethodGet)

	r.HandleFunc("/member", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleCreateMember, "admin", "staff"))).Methods(http.MethodPost)

	r.HandleFunc("/member/{memberID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleUpdateMember, "admin", "staff"))).Methods(http.MethodPut)

	r.HandleFunc("/member/{memberID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleDeleteMember, "admin", "staff"))).Methods(http.MethodDelete)
}

func (h *Handler) handleGetMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := utils.ParseStringToInt(r.URL.Query().Get("page"))

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	members, lastPage, err := h.memberStore.GetMembersWithPagination(ctx, page)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:     cok,
		Data:     members,
		LastPage: lastPage,
		Page:     page,
		Status:   http.StatusText(cok),
	})
}

func (h *Handler) handleGetMemberByID(w http.ResponseWriter, r *http.Request) {
	memberID := mux.Vars(r)["memberID"]

	ctx := r.Context()

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := uuid.Validate(memberID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	member, err := h.memberStore.GetMemberByID(ctx, memberID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Data:   member,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleCreateMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var filename string

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := r.ParseMultipartForm(size1MB); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer r.MultipartForm.RemoveAll()

	payload := types.SetPayloadMember{
		Nama:         r.PostForm.Get("nama"),
		JenisKelamin: r.PostForm.Get("jenis_kelamin"),
		Kelas:        r.PostForm.Get("kelas"),
		NoTelepon:    r.PostForm.Get("no_telepon"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesian(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	if _, err := h.memberStore.GetMemberByNama(ctx, payload.Nama); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("nama: %v has already exist", payload.Nama))
		return
	}

	if _, err := h.memberStore.GetMemberByNoTelepon(ctx, payload.NoTelepon); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("no_telepon: %v has already exist", payload.NoTelepon))
		return
	}

	_, header, err := r.FormFile("profil")
	if err == http.ErrMissingFile {
		filename = "-"
	}

	if err == nil {
		filename, err = utils.SetNewFilenameImg(ctx, "original", header, r2MembersAvatarPath)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	err = h.memberStore.CreateMember(ctx, &types.Member{
		Nama:          payload.Nama,
		JenisKelamin:  payload.JenisKelamin,
		Kelas:         payload.Kelas,
		NoTelepon:     payload.NoTelepon,
		ProfilAnggota: filename,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonResponse{
		Code:    http.StatusCreated,
		Message: "Member Created!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateMember(w http.ResponseWriter, r *http.Request) {
	memberID := mux.Vars(r)["memberID"]

	ctx := r.Context()

	var filename string

	if r.Method != http.MethodPut {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := uuid.Validate(memberID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := r.ParseMultipartForm(size1MB); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer r.MultipartForm.RemoveAll()

	payload := types.SetPayloadUpdateMember{
		Nama:         r.PostForm.Get("nama"),
		JenisKelamin: r.PostForm.Get("jenis_kelamin"),
		Kelas:        r.PostForm.Get("kelas"),
		NoTelepon:    r.PostForm.Get("no_telepon"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesian(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	m, err := h.memberStore.GetMemberByID(ctx, memberID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if payload.Nama != "" {
		m.Nama = payload.Nama
	}
	if payload.JenisKelamin != "" {
		m.JenisKelamin = payload.JenisKelamin
	}
	if payload.Kelas != "" {
		m.Kelas = payload.Kelas
	}
	if payload.NoTelepon != "" {
		m.NoTelepon = payload.NoTelepon
	}

	_, header, err := r.FormFile("profil")

	if err == http.ErrMissingFile {
		filename = m.ProfilAnggota
	}

	if err == nil {
		filename, err = utils.UpdateTheFilenameImg(ctx, "original", header, r2MembersAvatarPath, m.ProfilAnggota)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	err = h.memberStore.UpdateMember(ctx, memberID, &types.Member{
		Nama:          m.Nama,
		JenisKelamin:  m.JenisKelamin,
		Kelas:         m.Kelas,
		NoTelepon:     m.NoTelepon,
		ProfilAnggota: filename,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "Member Updated!",
		Status:  http.StatusText(cok),
	})
}

func (h *Handler) handleDeleteMember(w http.ResponseWriter, r *http.Request) {
	memberID := mux.Vars(r)["memberID"]

	ctx := r.Context()

	if r.Method != http.MethodDelete {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := uuid.Validate(memberID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	m, err := h.memberStore.GetMemberByID(ctx, memberID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	if err := utils.DeleteFilepathWithFilename(ctx, m.ProfilAnggota); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.memberStore.DeleteMember(ctx, memberID); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "Member Deleted!",
		Status:  http.StatusText(cok),
	})
}
