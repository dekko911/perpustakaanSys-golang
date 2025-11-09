package api

import (
	"database/sql"
	"net/http"
	"perpus_backend/pkg/cookie"
	"perpus_backend/pkg/cors"
	"perpus_backend/pkg/jwt"
	"perpus_backend/pkg/limiter"
	"perpus_backend/service/auth"
	"perpus_backend/service/book"
	"perpus_backend/service/circulation"
	"perpus_backend/service/member"
	"perpus_backend/service/role"
	roleuser "perpus_backend/service/role_user"
	"perpus_backend/service/user"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

var publicURLHandler = http.StripPrefix("/public/", http.FileServer(http.Dir("./assets/public")))

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	r := mux.NewRouter()
	r.Use(cors.CORSMiddleware)
	r.Use(cookie.CookieMiddleware)

	// for ensures that OPTIONS "/" is not thrown to 404 (which does not have a CORS header).
	r.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// set access files to public url.
	r.PathPrefix("/public/").Methods(http.MethodGet).Handler(publicURLHandler)

	subrouter := r.PathPrefix("/api").Subrouter()
	subrouter.Use(limiter.RateLimitMiddleware(rate.Every(2*time.Minute), 100))

	// for ensures that OPTIONS "/api" is not thrown to 404 (which does not have a CORS header).
	subrouter.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// user routes
	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)

	// role routes
	roleStore := role.NewStore(s.db)
	roleHandler := role.NewHandler(roleStore, userStore)
	roleHandler.RegisterRoutes(subrouter)

	// role_user routes
	roleUserStore := roleuser.NewStore(s.db)
	roleUserHandler := roleuser.NewHandler(roleUserStore, userStore, roleStore)
	roleUserHandler.RegisterRoutes(subrouter)

	// book routes
	bookStore := book.NewStore(s.db)
	bookHandler := book.NewHandler(bookStore, userStore)
	bookHandler.RegisterRoutes(subrouter)

	// member routes
	memberStore := member.NewStore(s.db)
	memberHandler := member.NewHandler(memberStore, userStore)
	memberHandler.RegisterRoutes(subrouter)

	// circulation routes
	circulationStore := circulation.NewStore(s.db)
	circulationHandler := circulation.NewHandler(circulationStore, userStore)
	circulationHandler.RegisterRoutes(subrouter)

	// auth routes
	authHandler := auth.NewHandler(userStore)
	authHandler.RegisterRoutes(subrouter)

	// private routes
	r.HandleFunc("/private/{filename:.+}", jwt.AuthWithJWTToken(jwt.NeededRole(userStore, "admin", "staff", "user")(authHandler.PrivateURLHandler), userStore)).Methods(http.MethodGet)

	return http.ListenAndServe(s.addr, r)
}
