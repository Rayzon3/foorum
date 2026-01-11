package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"jabber_v3/apps/api/internal/auth"
	"jabber_v3/apps/api/internal/http/handlers"
	"jabber_v3/apps/api/internal/http/middleware"
	"jabber_v3/apps/api/internal/http/response"
	"jabber_v3/apps/api/internal/http/utils"
	"jabber_v3/apps/api/internal/rtc"
	"jabber_v3/apps/api/internal/store"
)

func NewRouter(store *store.Store, jwt auth.JWTManager, allowedOrigins []string) http.Handler {
	handler := handlers.New(store, jwt)
	rtcHandler := rtc.NewHandler(jwt, rtc.NewRoomManager())
	r := chi.NewRouter()
	r.Use(utils.RequestLogger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	roomsRouter := chi.NewRouter()
	roomsRouter.Get("/{roomID}/ws", func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "roomID")
		rtcHandler.ServeWS(w, r, roomID)
	})
	r.Mount("/api/v1/rooms", roomsRouter)

	apiRouter := chi.NewRouter()
	apiRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	apiRouter.Post("/auth/register", handler.HandleRegister)
	apiRouter.Post("/auth/login", handler.HandleLogin)
	apiRouter.With(middleware.Auth(jwt)).Get("/me", handler.HandleMe)

	apiRouter.Route("/posts", func(r chi.Router) {
		r.With(middleware.OptionalAuth(jwt)).Get("/", handler.HandleFeed)
		r.With(middleware.Auth(jwt)).Post("/", handler.HandleCreatePost)
		r.With(middleware.Auth(jwt)).Post("/{postID}/vote", handler.HandleVote)
	})

	r.Mount("/api/v1", apiRouter)

	return r
}
