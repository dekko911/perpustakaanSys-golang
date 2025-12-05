package member

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"perpus_backend/pkg/jwt"
	"perpus_backend/types"
	"perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.MemberStore
	userStore types.UserStore
}

const (
	cok = http.StatusOK // for alias http.StatusOK

	dirAvatarPath = "./assets/public/images/avatar/"

	size1MB = 1 << 20
)

func NewHandler(s types.MemberStore, us types.UserStore) *Handler {
	return &Handler{store: s, userStore: us}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/members", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleGetMembers), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/members/{memberID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleGetMemberByID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/members", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleCreateMember), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/members/{memberID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleUpdateMember), h.userStore)).Methods(http.MethodPut)

	r.HandleFunc("/members/{memberID}", jwt.AuthWithJWTToken(jwt.RoleGate(h.userStore, "admin", "staff")(h.handleDeleteMember), h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.store.GetMembers()
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:   cok,
		Data:   members,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleGetMemberByID(w http.ResponseWriter, r *http.Request) {
	memberID := mux.Vars(r)["memberID"]

	if err := uuid.Validate(memberID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	member, err := h.store.GetMemberByID(memberID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:   cok,
		Data:   member,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleCreateMember(w http.ResponseWriter, r *http.Request) {
	var (
		fileName, avatarPath, extFile string
		sizeFile                      int64
	)

	r.Body = http.MaxBytesReader(w, r.Body, size1MB)

	if err := r.ParseMultipartForm(size1MB); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	payload := types.SetPayloadMember{
		Nama:         r.FormValue("nama"),
		JenisKelamin: r.FormValue("jenis_kelamin"),
		Kelas:        r.FormValue("kelas"),
		NoTelepon:    r.FormValue("no_telepon"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	if _, err := h.store.GetMemberByNama(payload.Nama); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("nama: %v has already exist", payload.Nama))
		return
	}

	if _, err := h.store.GetMemberByNoTelepon(payload.NoTelepon); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("no_telepon: %v has already exist", payload.NoTelepon))
		return
	}

	file, header, err := r.FormFile("profil")
	if err == http.ErrMissingFile {
		fileName = "-"
	}

	if err == nil {
		defer file.Close()

		sizeFile = header.Size
		extFile = filepath.Ext(header.Filename)

		if extFile != ".png" && extFile != ".jpeg" && extFile != ".jpg" {
			utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("only support jpg, jpeg, and png"))
			return
		}

		if sizeFile > size1MB {
			utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("serve file under 1mb"))
			return
		}

		fileName = header.Filename
		avatarPath = dirAvatarPath + fileName

		dst, _ := os.Create(avatarPath)
		defer dst.Close()

		io.Copy(dst, file)
	}

	if err := h.store.CreateMember(&types.Member{
		Nama:          payload.Nama,
		JenisKelamin:  payload.JenisKelamin,
		Kelas:         payload.Kelas,
		NoTelepon:     payload.NoTelepon,
		ProfilAnggota: fileName,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonData{
		Code:    http.StatusCreated,
		Message: "Member Created!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateMember(w http.ResponseWriter, r *http.Request) {
	memberID := mux.Vars(r)["memberID"]

	var (
		fileName, avatarPath, extFile string
		sizeFile                      int64
	)

	if r.Method != http.MethodPut {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, size1MB)

	if err := uuid.Validate(memberID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := r.ParseMultipartForm(size1MB); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	payload := types.SetPayloadUpdateMember{
		Nama:         r.FormValue("nama"),
		JenisKelamin: r.FormValue("jenis_kelamin"),
		Kelas:        r.FormValue("kelas"),
		NoTelepon:    r.FormValue("no_telepon"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	m, err := h.store.GetMemberByID(memberID)
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

	file, header, err := r.FormFile("profil")
	if err == http.ErrMissingFile {
		fileName = m.ProfilAnggota
	}

	if err == nil {
		defer file.Close()

		sizeFile = header.Size
		extFile = filepath.Ext(header.Filename)

		if extFile != ".png" && extFile != ".jpg" && extFile != ".jpeg" {
			utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only supports jpg, jpeg, and png"))
			return
		}

		if sizeFile > size1MB {
			utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only serve file under 1mb"))
			return
		}

		avatarPathOld := dirAvatarPath + m.ProfilAnggota
		info, err := os.Stat(avatarPathOld)

		if err == nil {
			if !info.IsDir() {
				os.Remove(avatarPathOld)
			}
		}

		fileName = header.Filename
		avatarPath = dirAvatarPath + fileName

		dst, _ := os.Create(avatarPath)
		defer dst.Close()

		io.Copy(dst, file)
	}

	if err := h.store.UpdateMember(memberID, &types.Member{
		Nama:          m.Nama,
		JenisKelamin:  m.JenisKelamin,
		Kelas:         m.Kelas,
		NoTelepon:     m.NoTelepon,
		ProfilAnggota: fileName,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:    cok,
		Message: "Member Updated!",
		Status:  http.StatusText(cok),
	})
}

func (h *Handler) handleDeleteMember(w http.ResponseWriter, r *http.Request) {
	memberID := mux.Vars(r)["memberID"]

	if err := uuid.Validate(memberID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	m, err := h.store.GetMemberByID(memberID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, err)
		return
	}

	filePath := dirAvatarPath + m.ProfilAnggota
	info, err := os.Stat(filePath)

	if err == nil {
		if !info.IsDir() {
			os.Remove(filePath)
		}
	}

	if err := h.store.DeleteMember(memberID); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:    cok,
		Message: "Member Deleted!",
		Status:  http.StatusText(cok),
	})
}
