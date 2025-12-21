package auth

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/perpus_backend/pkg/hash"
	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

type Handler struct {
	store types.UserStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, store types.UserStore) *Handler {
	return &Handler{store: store, jwt: jwt}
}

const (
	cok = http.StatusOK

	profilePath = "./assets/public/images/profile/"
	privateDir  = "./assets/private"

	size1MB = 1 << 20
)

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/login", h.handleLogin).Methods(http.MethodPost)
	r.HandleFunc("/register", h.handleRegister).Methods(http.MethodPost)
	r.HandleFunc("/logout", h.jwt.AuthWithJWTToken(h.handleLogout)).Methods(http.MethodPost)
}

// Handler auth login using JWT.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("method post only"))
		return
	}

	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadLogin{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	u, err := h.store.GetUserWithRolesByEmail(ctx, payload.Email)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("wrong email"))
		return
	}

	if !hash.CompareHashedPassword(u.Password, []byte(payload.Password)) {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("wrong password"))
		return
	}

	token, err := h.jwt.CreateTokenJWT(ctx, u.ID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:   cok,
		Status: http.StatusText(cok),
		Token:  token,
	})
}

// Handle Logout and Revoke the Token using token versioning (user session ver).
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := jwt.GetUserIDFromContext(ctx)

	token := utils.GetTokenFromRequest(r)

	err := h.store.IncrementTokenVersion(ctx, userID, token)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:    cok,
		Message: "You've been Logout!",
		Status:  http.StatusText(cok),
	})
}

// Handle register user, this not will add the role.
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()

		fileName string
	)

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("method post only"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, size1MB)

	if err := r.ParseMultipartForm(size1MB); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload := types.SetPayloadUser{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	file, header, errFile := r.FormFile("avatar")

	if errFile == http.ErrMissingFile {
		fileName = "-"
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	if _, err := h.store.GetUserWithRolesByEmail(ctx, payload.Email); err == nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashPass, err := hash.HashPassword(payload.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if errFile == nil {
		defer file.Close()

		randomString := xid.New().String()
		ext := filepath.Ext(header.Filename)

		fileName = randomString + ext

		if size1MB <= header.Size {
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				dst, _ := os.Create(profilePath + fileName)
				defer dst.Close()

				io.Copy(dst, file)
			} else {
				utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("only serve png, jpg, and jpeg file"))
				return
			}
		} else {
			utils.WriteJSONError(w, http.StatusForbidden, fmt.Errorf("only serve file under 1 mb"))
			return
		}
	}

	if err := h.store.CreateUser(ctx, &types.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hashPass,
		Avatar:   fileName,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonData{
		Code:    http.StatusCreated,
		Message: "User Registered!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) PrivateURLHandler(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["filename"]

	joined := filepath.Join(privateDir, filename)
	cleaned := filepath.Clean(joined)

	if !utils.IsItInBaseDir(cleaned, privateDir) {
		utils.WriteJSONError(w, http.StatusNotFound, fmt.Errorf("file not found"))
		return
	}

	http.ServeFile(w, r, cleaned)
}
