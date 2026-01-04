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
  "jabber_v3/apps/api/internal/store"
)

func NewRouter(store *store.Store, jwt auth.JWTManager, allowedOrigins []string) http.Handler {
  handler := handlers.New(store, jwt)
  r := chi.NewRouter()
  r.Use(utils.RequestLogger)
  r.Use(cors.Handler(cors.Options{
    AllowedOrigins: allowedOrigins,
    AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
    AllowCredentials: false,
    MaxAge: 300,
  }))

  r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
  })

  r.Route("/api/v1", func(r chi.Router) {
    r.Post("/auth/register", handler.HandleRegister)
    r.Post("/auth/login", handler.HandleLogin)
    r.With(middleware.Auth(jwt)).Get("/me", handler.HandleMe)

    r.Route("/posts", func(r chi.Router) {
      r.With(middleware.OptionalAuth(jwt)).Get("/", handler.HandleFeed)
      r.With(middleware.Auth(jwt)).Post("/", handler.HandleCreatePost)
      r.With(middleware.Auth(jwt)).Post("/{postID}/vote", handler.HandleVote)
    })
  })

  return r
}
