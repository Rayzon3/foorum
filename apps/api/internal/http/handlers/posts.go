package handlers

import (
  "encoding/json"
  "net/http"
  "strings"

  "github.com/go-chi/chi/v5"

  "jabber_v3/apps/api/internal/http/requestctx"
  "jabber_v3/apps/api/internal/http/response"
  "jabber_v3/apps/api/internal/http/types"
  "jabber_v3/apps/api/internal/store"
)

func (h *Handler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
  user, ok := requestctx.AuthUserFromContext(r.Context())
  if !ok {
    response.WriteError(w, http.StatusUnauthorized, "unauthorized")
    return
  }

  record, err := h.store.Users.GetUserByID(r.Context(), user.ID)
  if err != nil {
    if err == store.ErrNotFound {
      response.WriteError(w, http.StatusUnauthorized, "unauthorized")
      return
    }
    response.WriteError(w, http.StatusInternalServerError, "fetch_failed")
    return
  }

  var req types.CreatePostRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response.WriteError(w, http.StatusBadRequest, "invalid_json")
    return
  }

  title := strings.TrimSpace(req.Title)
  body := strings.TrimSpace(req.Body)
  if title == "" || body == "" {
    response.WriteError(w, http.StatusBadRequest, "invalid_post")
    return
  }

  post, err := h.store.Posts.CreatePost(r.Context(), record.ID, title, body)
  if err != nil {
    response.WriteError(w, http.StatusInternalServerError, "create_failed")
    return
  }

  response.WriteJSON(w, http.StatusCreated, types.PostView{
    ID: post.ID,
    Title: post.Title,
    Body: post.Body,
    CreatedAt: post.CreatedAt,
    Author: types.AuthorView{ID: record.ID, Email: record.Email, Username: record.Username},
    Score: 0,
    MyVote: 0,
  })
}

func (h *Handler) HandleFeed(w http.ResponseWriter, r *http.Request) {
  var userID *string
  if user, ok := requestctx.AuthUserFromContext(r.Context()); ok {
    userID = &user.ID
  }

  posts, err := h.store.Posts.ListFeed(r.Context(), 50, userID)
  if err != nil {
    response.WriteError(w, http.StatusInternalServerError, "fetch_failed")
    return
  }

  views := make([]types.PostView, 0, len(posts))
  for _, post := range posts {
    views = append(views, types.PostView{
      ID: post.ID,
      Title: post.Title,
      Body: post.Body,
      CreatedAt: post.CreatedAt,
      Author: types.AuthorView{
        ID: post.UserID,
        Email: post.AuthorEmail,
        Username: post.AuthorUsername,
      },
      Score: post.Score,
      MyVote: post.MyVote,
    })
  }

  response.WriteJSON(w, http.StatusOK, views)
}

func (h *Handler) HandleVote(w http.ResponseWriter, r *http.Request) {
  user, ok := requestctx.AuthUserFromContext(r.Context())
  if !ok {
    response.WriteError(w, http.StatusUnauthorized, "unauthorized")
    return
  }

  postID := strings.TrimSpace(chi.URLParam(r, "postID"))
  if postID == "" {
    response.WriteError(w, http.StatusBadRequest, "invalid_post")
    return
  }

  var req types.VoteRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response.WriteError(w, http.StatusBadRequest, "invalid_json")
    return
  }

  if req.Value == 0 {
    if err := h.store.Votes.DeleteVote(r.Context(), postID, user.ID); err != nil {
      response.WriteError(w, http.StatusInternalServerError, "vote_failed")
      return
    }
    response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
    return
  }

  if req.Value != 1 && req.Value != -1 {
    response.WriteError(w, http.StatusBadRequest, "invalid_vote")
    return
  }

  if err := h.store.Votes.UpsertVote(r.Context(), postID, user.ID, req.Value); err != nil {
    response.WriteError(w, http.StatusInternalServerError, "vote_failed")
    return
  }

  response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
