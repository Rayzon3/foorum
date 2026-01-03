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

func (s *UserStore) CreateUser(ctx context.Context, email string, username string, passwordHash string) (models.User, error) {
  var user models.User
  err := s.db.QueryRowContext(
    ctx,
    "insert into users (email, username, password_hash) values ($1, $2, $3) returning id, email, username, password_hash, created_at",
    email,
    username,
    passwordHash,
  ).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
  var user models.User
  err := s.db.QueryRowContext(
    ctx,
    "select id, email, username, password_hash, created_at from users where email = $1",
    email,
  ).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
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
    "select id, email, username, password_hash, created_at from users where id = $1",
    id,
  ).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
  if errors.Is(err, sql.ErrNoRows) {
    return models.User{}, ErrNotFound
  }
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}

func (s *UserStore) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
  var user models.User
  err := s.db.QueryRowContext(
    ctx,
    "select id, email, username, password_hash, created_at from users where username = $1",
    username,
  ).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
  if errors.Is(err, sql.ErrNoRows) {
    return models.User{}, ErrNotFound
  }
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}
