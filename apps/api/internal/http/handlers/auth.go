package handlers

import (
  "encoding/json"
  "errors"
  "net/http"
  "strings"

  "github.com/jackc/pgx/v5/pgconn"

  "jabber_v3/apps/api/internal/auth"
  "jabber_v3/apps/api/internal/http/response"
  "jabber_v3/apps/api/internal/http/types"
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
  if email == "" || len(req.Password) < 8 {
    response.WriteError(w, http.StatusBadRequest, "invalid_credentials")
    return
  }

  hash, err := auth.HashPassword(req.Password)
  if err != nil {
    response.WriteError(w, http.StatusInternalServerError, "hash_failed")
    return
  }

  user, err := h.store.Users.CreateUser(r.Context(), email, hash)
  if err != nil {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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
    User: types.UserView{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt},
  })
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
  var req types.CredentialsRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response.WriteError(w, http.StatusBadRequest, "invalid_json")
    return
  }

  email := normalizeEmail(req.Email)
  if email == "" || req.Password == "" {
    response.WriteError(w, http.StatusBadRequest, "invalid_credentials")
    return
  }

  user, err := h.store.Users.GetUserByEmail(r.Context(), email)
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
    User: types.UserView{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt},
  })
}

func normalizeEmail(email string) string {
  email = strings.TrimSpace(strings.ToLower(email))
  return email
}
