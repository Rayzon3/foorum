package store

import (
  "context"
  "database/sql"
  "errors"

  "jabber_v3/apps/api/internal/models"
  "jabber_v3/apps/api/internal/store/qb"
)

type PostStore struct {
  db *sql.DB
}

var postColumns = []string{"id", "user_id", "title", "body", "created_at"}

func (s *PostStore) CreatePost(ctx context.Context, userID string, title string, body string) (models.Post, error) {
  query, args := qb.Insert("posts").
    Columns("user_id", "title", "body").
    Values(userID, title, body).
    Returning(postColumns...).
    Build()
  row := s.db.QueryRowContext(ctx, query, args...)
  post, err := scanPost(row)
  if err != nil {
    return models.Post{}, err
  }
  return post, nil
}

func (s *PostStore) ListFeed(ctx context.Context, limit int, userID *string) ([]models.PostWithStats, error) {
  var args []any
  query := `
    select
      p.id,
      p.user_id,
      p.title,
      p.body,
      p.created_at,
      u.email,
      u.username,
      coalesce(sum(v.value), 0) as score,
      coalesce(max(case when v.user_id = $1 then v.value end), 0) as my_vote
    from posts p
    join users u on u.id = p.user_id
    left join post_votes v on v.post_id = p.id
    group by p.id, u.email, u.username
    order by p.created_at desc
    limit $2`
  viewerID := ""
  if userID != nil {
    viewerID = *userID
  }
  args = append(args, viewerID, limit)

  rows, err := s.db.QueryContext(ctx, query, args...)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var posts []models.PostWithStats
  for rows.Next() {
    post, err := scanPostWithStats(rows)
    if err != nil {
      return nil, err
    }
    posts = append(posts, post)
  }
  if err := rows.Err(); err != nil {
    return nil, err
  }
  return posts, nil
}

func scanPost(row rowScanner) (models.Post, error) {
  var post models.Post
  err := row.Scan(&post.ID, &post.UserID, &post.Title, &post.Body, &post.CreatedAt)
  if err != nil {
    return models.Post{}, err
  }
  return post, nil
}

func scanPostWithStats(row rowScanner) (models.PostWithStats, error) {
  var post models.PostWithStats
  err := row.Scan(
    &post.ID,
    &post.UserID,
    &post.Title,
    &post.Body,
    &post.CreatedAt,
    &post.AuthorEmail,
    &post.AuthorUsername,
    &post.Score,
    &post.MyVote,
  )
  if err != nil {
    return models.PostWithStats{}, err
  }
  return post, nil
}

func (s *PostStore) GetPostByID(ctx context.Context, postID string) (models.Post, error) {
  query, args := qb.Select(postColumns...).
    From("posts").
    WhereEq("id", postID).
    Build()
  row := s.db.QueryRowContext(ctx, query, args...)
  post, err := scanPost(row)
  if errors.Is(err, sql.ErrNoRows) {
    return models.Post{}, ErrNotFound
  }
  if err != nil {
    return models.Post{}, err
  }
  return post, nil
}
