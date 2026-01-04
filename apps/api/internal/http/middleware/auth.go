package middleware

import (
  "net/http"
  "strings"

  "jabber_v3/apps/api/internal/auth"
  "jabber_v3/apps/api/internal/http/requestctx"
  "jabber_v3/apps/api/internal/http/response"
  "jabber_v3/apps/api/internal/http/types"
)

func Auth(jwt auth.JWTManager) func(http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      header := r.Header.Get("Authorization")
      if header == "" {
        response.WriteError(w, http.StatusUnauthorized, "unauthorized")
        return
      }

      parts := strings.SplitN(header, " ", 2)
      if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
        response.WriteError(w, http.StatusUnauthorized, "unauthorized")
        return
      }

      claims, err := jwt.Parse(parts[1])
      if err != nil {
        response.WriteError(w, http.StatusUnauthorized, "unauthorized")
        return
      }

      ctx := requestctx.WithAuthUser(r.Context(), types.AuthUser{ID: claims.UserID})
      next.ServeHTTP(w, r.WithContext(ctx))
    })
  }
}

func OptionalAuth(jwt auth.JWTManager) func(http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      header := r.Header.Get("Authorization")
      if header == "" {
        next.ServeHTTP(w, r)
        return
      }

      parts := strings.SplitN(header, " ", 2)
      if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
        next.ServeHTTP(w, r)
        return
      }

      claims, err := jwt.Parse(parts[1])
      if err != nil {
        next.ServeHTTP(w, r)
        return
      }

      ctx := requestctx.WithAuthUser(r.Context(), types.AuthUser{ID: claims.UserID})
      next.ServeHTTP(w, r.WithContext(ctx))
    })
  }
}
