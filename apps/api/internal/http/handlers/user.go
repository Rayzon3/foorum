package handlers

import (
  "errors"
  "net/http"

  "jabber_v3/apps/api/internal/http/requestctx"
  "jabber_v3/apps/api/internal/http/response"
  "jabber_v3/apps/api/internal/http/types"
  "jabber_v3/apps/api/internal/store"
)

func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
  user, ok := requestctx.AuthUserFromContext(r.Context())
  if !ok {
    response.WriteError(w, http.StatusUnauthorized, "unauthorized")
    return
  }

  record, err := h.store.Users.GetUserByID(r.Context(), user.ID)
  if err != nil {
    if errors.Is(err, store.ErrNotFound) {
      response.WriteError(w, http.StatusUnauthorized, "unauthorized")
      return
    }
    response.WriteError(w, http.StatusInternalServerError, "fetch_failed")
    return
  }

  response.WriteJSON(w, http.StatusOK, types.UserView{ID: record.ID, Email: record.Email, Username: record.Username, CreatedAt: record.CreatedAt})
}
