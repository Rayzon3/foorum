package requestctx

import (
  "context"

  "jabber_v3/apps/api/internal/http/types"
)

type authUserKey struct{}

func WithAuthUser(ctx context.Context, user types.AuthUser) context.Context {
  return context.WithValue(ctx, authUserKey{}, user)
}

func AuthUserFromContext(ctx context.Context) (types.AuthUser, bool) {
  user, ok := ctx.Value(authUserKey{}).(types.AuthUser)
  return user, ok
}
