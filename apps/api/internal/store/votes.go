package store

import (
  "context"
  "database/sql"

  "jabber_v3/apps/api/internal/store/qb"
)

type VoteStore struct {
  db *sql.DB
}

func (s *VoteStore) UpsertVote(ctx context.Context, postID string, userID string, value int) error {
  query := `
    insert into post_votes (post_id, user_id, value)
    values ($1, $2, $3)
    on conflict (post_id, user_id)
    do update set value = excluded.value`
  _, err := s.db.ExecContext(ctx, query, postID, userID, value)
  return err
}

func (s *VoteStore) DeleteVote(ctx context.Context, postID string, userID string) error {
  query, args := qb.Delete("post_votes").
    WhereEq("post_id", postID).
    WhereEq("user_id", userID).
    Build()
  _, err := s.db.ExecContext(ctx, query, args...)
  return err
}
