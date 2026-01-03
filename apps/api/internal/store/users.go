package store

import (
  "context"
  "database/sql"
  "errors"

  "jabber_v3/apps/api/internal/models"
)

type UserStore struct {
  db *sql.DB
}

func (s *UserStore) CreateUser(ctx context.Context, email string, passwordHash string) (models.User, error) {
  var user models.User
  err := s.db.QueryRowContext(
    ctx,
    "insert into users (email, password_hash) values ($1, $2) returning id, email, password_hash, created_at",
    email,
    passwordHash,
  ).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
  var user models.User
  err := s.db.QueryRowContext(
    ctx,
    "select id, email, password_hash, created_at from users where email = $1",
    email,
  ).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
  if errors.Is(err, sql.ErrNoRows) {
    return models.User{}, ErrNotFound
  }
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}

func (s *UserStore) GetUserByID(ctx context.Context, id string) (models.User, error) {
  var user models.User
  err := s.db.QueryRowContext(
    ctx,
    "select id, email, password_hash, created_at from users where id = $1",
    id,
  ).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
  if errors.Is(err, sql.ErrNoRows) {
    return models.User{}, ErrNotFound
  }
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}
