package httpapi

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "net/http"
  "strings"
  "time"

  "github.com/go-chi/chi/v5"
  "github.com/go-chi/cors"
  "github.com/jackc/pgx/v5/pgconn"

  "jabber_v3/apps/api/internal/auth"
  "jabber_v3/apps/api/internal/store"
)

type Server struct {
  store *store.Store
  jwt   auth.JWTManager
}

type authContextKey struct{}

type AuthUser struct {
  ID    string
  Email string
}

func NewServer(store *store.Store, jwt auth.JWTManager) *Server {
  return &Server{store: store, jwt: jwt}
}

func (s *Server) Routes(allowedOrigins []string) http.Handler {
  r := chi.NewRouter()
  r.Use(requestLogger)
  r.Use(cors.Handler(cors.Options{
    AllowedOrigins: allowedOrigins,
    AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
    AllowCredentials: false,
    MaxAge: 300,
  }))

  r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
  })

  r.Route("/api/v1", func(r chi.Router) {
    r.Post("/auth/register", s.handleRegister)
    r.Post("/auth/login", s.handleLogin)
    r.With(s.authMiddleware).Get("/me", s.handleMe)
  })

  return r
}

type statusRecorder struct {
  http.ResponseWriter
  status int
}

func (sr *statusRecorder) WriteHeader(code int) {
  sr.status = code
  sr.ResponseWriter.WriteHeader(code)
}

func requestLogger(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
    next.ServeHTTP(sr, r)
    durationMs := time.Since(start).Milliseconds()
    log.Printf("method=%s path=%s status=%s duration_ms=%d", r.Method, r.URL.Path, colorStatus(sr.status), durationMs)
  })
}

func colorStatus(status int) string {
  color := "\x1b[32m" // green
  switch {
  case status >= 500:
    color = "\x1b[31m" // red
  case status >= 400:
    color = "\x1b[33m" // yellow
  case status >= 300:
    color = "\x1b[36m" // cyan
  }
  return fmt.Sprintf("%s%d\x1b[0m", color, status)
}

type credentialsRequest struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}

type authResponse struct {
  Token string    `json:"token"`
  User  userView  `json:"user"`
}

type userView struct {
  ID        string    `json:"id"`
  Email     string    `json:"email"`
  CreatedAt time.Time `json:"createdAt"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
  var req credentialsRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    writeError(w, http.StatusBadRequest, "invalid_json")
    return
  }

  email := normalizeEmail(req.Email)
  if email == "" || len(req.Password) < 8 {
    writeError(w, http.StatusBadRequest, "invalid_credentials")
    return
  }

  hash, err := auth.HashPassword(req.Password)
  if err != nil {
    writeError(w, http.StatusInternalServerError, "hash_failed")
    return
  }

  user, err := s.store.CreateUser(r.Context(), email, hash)
  if err != nil {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
      writeError(w, http.StatusConflict, "email_taken")
      return
    }
    writeError(w, http.StatusInternalServerError, "create_failed")
    return
  }

  token, err := s.jwt.Generate(user.ID)
  if err != nil {
    writeError(w, http.StatusInternalServerError, "token_failed")
    return
  }

  writeJSON(w, http.StatusCreated, authResponse{
    Token: token,
    User: userView{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt},
  })
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
  var req credentialsRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    writeError(w, http.StatusBadRequest, "invalid_json")
    return
  }

  email := normalizeEmail(req.Email)
  if email == "" || req.Password == "" {
    writeError(w, http.StatusBadRequest, "invalid_credentials")
    return
  }

  user, err := s.store.GetUserByEmail(r.Context(), email)
  if err != nil {
    if errors.Is(err, store.ErrNotFound) {
      writeError(w, http.StatusUnauthorized, "invalid_login")
      return
    }
    writeError(w, http.StatusInternalServerError, "login_failed")
    return
  }

  if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
    writeError(w, http.StatusUnauthorized, "invalid_login")
    return
  }

  token, err := s.jwt.Generate(user.ID)
  if err != nil {
    writeError(w, http.StatusInternalServerError, "token_failed")
    return
  }

  writeJSON(w, http.StatusOK, authResponse{
    Token: token,
    User: userView{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt},
  })
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
  authUser := r.Context().Value(authContextKey{})
  if authUser == nil {
    writeError(w, http.StatusUnauthorized, "unauthorized")
    return
  }

  user := authUser.(AuthUser)
  record, err := s.store.GetUserByID(r.Context(), user.ID)
  if err != nil {
    if errors.Is(err, store.ErrNotFound) {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }
    writeError(w, http.StatusInternalServerError, "fetch_failed")
    return
  }

  writeJSON(w, http.StatusOK, userView{ID: record.ID, Email: record.Email, CreatedAt: record.CreatedAt})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    header := r.Header.Get("Authorization")
    if header == "" {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    parts := strings.SplitN(header, " ", 2)
    if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    claims, err := s.jwt.Parse(parts[1])
    if err != nil {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    ctx := context.WithValue(r.Context(), authContextKey{}, AuthUser{ID: claims.UserID})
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}

func normalizeEmail(email string) string {
  email = strings.TrimSpace(strings.ToLower(email))
  return email
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  _ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string) {
  writeJSON(w, status, map[string]string{"error": code})
}
