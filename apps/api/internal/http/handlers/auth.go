package handlers

import (
  "context"
  "encoding/json"
  "errors"
  "net/http"
  "strings"

  "github.com/jackc/pgx/v5/pgconn"

  "jabber_v3/apps/api/internal/auth"
  "jabber_v3/apps/api/internal/http/response"
  "jabber_v3/apps/api/internal/http/types"
  "jabber_v3/apps/api/internal/models"
  "jabber_v3/apps/api/internal/store"
)

type Handler struct {
  store *store.Store
  jwt   auth.JWTManager
}

func New(store *store.Store, jwt auth.JWTManager) *Handler {
  return &Handler{store: store, jwt: jwt}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
  var req types.CredentialsRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response.WriteError(w, http.StatusBadRequest, "invalid_json")
    return
  }

  email := normalizeEmail(req.Email)
  username := normalizeUsername(req.Username)
  if email == "" || username == "" || len(req.Password) < 8 {
    response.WriteError(w, http.StatusBadRequest, "invalid_credentials")
    return
  }

  hash, err := auth.HashPassword(req.Password)
  if err != nil {
    response.WriteError(w, http.StatusInternalServerError, "hash_failed")
    return
  }

  user, err := h.store.Users.CreateUser(r.Context(), email, username, hash)
  if err != nil {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
      if pgErr.ConstraintName == "users_username_unique" {
        response.WriteError(w, http.StatusConflict, "username_taken")
        return
      }
      response.WriteError(w, http.StatusConflict, "email_taken")
      return
    }
    response.WriteError(w, http.StatusInternalServerError, "create_failed")
    return
  }

  token, err := h.jwt.Generate(user.ID)
  if err != nil {
    response.WriteError(w, http.StatusInternalServerError, "token_failed")
    return
  }

  response.WriteJSON(w, http.StatusCreated, types.AuthResponse{
    Token: token,
    User: types.UserView{ID: user.ID, Email: user.Email, Username: user.Username, CreatedAt: user.CreatedAt},
  })
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
  var req types.CredentialsRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response.WriteError(w, http.StatusBadRequest, "invalid_json")
    return
  }

  identifier := strings.TrimSpace(req.Email)
  if identifier == "" {
    identifier = strings.TrimSpace(req.Username)
  }
  if identifier == "" || req.Password == "" {
    response.WriteError(w, http.StatusBadRequest, "invalid_credentials")
    return
  }

  user, err := h.lookupUser(r.Context(), identifier)
  if err != nil {
    if errors.Is(err, store.ErrNotFound) {
      response.WriteError(w, http.StatusUnauthorized, "invalid_login")
      return
    }
    response.WriteError(w, http.StatusInternalServerError, "login_failed")
    return
  }

  if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
    response.WriteError(w, http.StatusUnauthorized, "invalid_login")
    return
  }

  token, err := h.jwt.Generate(user.ID)
  if err != nil {
    response.WriteError(w, http.StatusInternalServerError, "token_failed")
    return
  }

  response.WriteJSON(w, http.StatusOK, types.AuthResponse{
    Token: token,
    User: types.UserView{ID: user.ID, Email: user.Email, Username: user.Username, CreatedAt: user.CreatedAt},
  })
}

func normalizeEmail(email string) string {
  email = strings.TrimSpace(strings.ToLower(email))
  return email
}

func normalizeUsername(username string) string {
  username = strings.TrimSpace(strings.ToLower(username))
  return username
}

func (h *Handler) lookupUser(ctx context.Context, identifier string) (models.User, error) {
  if strings.Contains(identifier, "@") {
    return h.store.Users.GetUserByEmail(ctx, normalizeEmail(identifier))
  }
  return h.store.Users.GetUserByUsername(ctx, normalizeUsername(identifier))
}
