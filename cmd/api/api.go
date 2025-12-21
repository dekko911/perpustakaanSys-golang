package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/perpus_backend/config"
	"github.com/perpus_backend/pkg/cookie"
	"github.com/perpus_backend/pkg/cors"
	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/pkg/limiter"
	"github.com/perpus_backend/service/auth"
	"github.com/perpus_backend/service/book"
	"github.com/perpus_backend/service/circulation"
	"github.com/perpus_backend/service/member"
	"github.com/perpus_backend/service/role"
	roleuser "github.com/perpus_backend/service/role_user"
	"github.com/perpus_backend/service/user"
	"github.com/perpus_backend/service/websocket"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type APIServer struct {
	addr string

	db  *sql.DB
	rdb *redis.Client
}

func NewAPIServer(addr string, db *sql.DB, rdb *redis.Client) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
		rdb:  rdb,
	}
}

var publicURLHandler = http.StripPrefix("/public/", http.FileServer(http.Dir("./assets/public")))

func (s *APIServer) Run() error {
	r := mux.NewRouter()
	r.Use(cookie.CookieMiddleware)
	r.Use(cors.CORSMiddleware)

	// for ensures that OPTIONS "/" is not thrown to 404 (which does not have a CORS header).
	r.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	subrouter := r.PathPrefix("/api").Subrouter()

	// limiter for env production
	if config.Env.AppENV == "production" {
		r.Use(limiter.SetRateLimitMiddleware(rate.Every(1*time.Hour), 3000))
		subrouter.Use(limiter.SetRateLimitMiddleware(rate.Every(1*time.Minute), 10))
	}

	// for ensures that OPTIONS "/api" is not thrown to 404 (which does not have a CORS header).
	subrouter.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	userStore := user.NewStore(s.db, s.rdb)

	jwt := jwt.NewAuthJWT(userStore, s.rdb)

	// user routes
	userHandler := user.NewHandler(jwt, userStore)
	userHandler.RegisterRoutes(subrouter)

	// role routes
	roleStore := role.NewStore(s.db, s.rdb)
	roleHandler := role.NewHandler(jwt, roleStore, userStore)
	roleHandler.RegisterRoutes(subrouter)

	// role_user routes
	roleUserStore := roleuser.NewStore(s.db, s.rdb)
	roleUserHandler := roleuser.NewHandler(jwt, roleUserStore, userStore, roleStore)
	roleUserHandler.RegisterRoutes(subrouter)

	// book routes
	bookStore := book.NewStore(s.db, s.rdb)
	bookHandler := book.NewHandler(jwt, bookStore, userStore)
	bookHandler.RegisterRoutes(subrouter)

	// circulation routes
	circulationStore := circulation.NewStore(s.db, s.rdb)
	circulationHandler := circulation.NewHandler(jwt, circulationStore, userStore)
	circulationHandler.RegisterRoutes(subrouter)

	// member routes
	memberStore := member.NewStore(s.db, s.rdb)
	memberHandler := member.NewHandler(jwt, memberStore, userStore)
	memberHandler.RegisterRoutes(subrouter)

	// auth routes
	authHandler := auth.NewHandler(jwt, userStore)
	authHandler.RegisterRoutes(subrouter)

	// search routes
	wsSubrouter := r.PathPrefix("/ws").Subrouter()
	wsHandler := websocket.NewHandler(jwt, userStore, roleStore, memberStore, bookStore, circulationStore)
	wsHandler.RegisterRoutes(wsSubrouter)

	r.PathPrefix("/public/").Handler(publicURLHandler).Methods(http.MethodGet) // set accessing files across public url.

	// get info logged profile
	r.HandleFunc("/profile", jwt.AuthWithJWTToken(userHandler.HandleGetProfileUser)).Methods(http.MethodGet)

	// set accessing files across private routes. Which means, it is need to login auth.
	r.HandleFunc("/private/{filename:.+}", jwt.AuthWithJWTToken(jwt.RoleGate(authHandler.PrivateURLHandler, "admin", "staff", "user"))).Methods(http.MethodGet)

	return http.ListenAndServe(s.addr, r)
}
