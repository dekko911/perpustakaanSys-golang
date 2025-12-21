package roleuser

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"

	"github.com/gorilla/mux"
)

func TestHandlerRoleUser(t *testing.T) {
	jwt := &jwt.AuthJWT{}
	roleUser := &types.MockRoleUserStore{}
	user := &types.MockUserStore{}
	role := &types.MockRoleStore{}

	h := NewHandler(jwt, roleUser, user, role)

	t.Run("it should be get user & role by userid", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/role_user/6918315b-dff4-8324-969f-e43cd434eb3e", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/role_user/{userID}", h.handleGetUserWithRoleByUserID).Methods(http.MethodGet)

		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for check the body purpose

		if w.Code != cok {
			t.Errorf("expected status code %d, got %d", cok, w.Code)
		}
	})

	// t.Run("it should be create relation between user & role", func(t *testing.T) {
	// 	form := url.Values{}

	// 	payload := types.SetPayloadRoleAndUserID{
	// 		UserID: uuid.NewString(),
	// 		RoleID: uuid.NewString(),
	// 	}

	// 	form.Set("user_id", payload.UserID)
	// 	form.Set("role_id", payload.RoleID)

	// 	req, err := http.NewRequest(http.MethodPost, "/role_user", strings.NewReader(form.Encode()))
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 	w := httptest.NewRecorder()
	// 	r := mux.NewRouter()

	// 	r.HandleFunc("/role_user", h.handleAssignRoleIntoUser).Methods(http.MethodPost)
	// 	r.ServeHTTP(w, req)

	// 	t.Log(w.Body)

	// 	if w.Code != COK {
	// 		t.Errorf("expected status code %d, got %d", COK, w.Code)
	// 	}
	// })
}
