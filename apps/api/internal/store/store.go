package store

import (
  "context"
  "database/sql"
  "errors"
  "time"
)

var ErrNotFound = errors.New("not found")

type Store struct {
  DB *sql.DB
}

type User struct {
  ID           string
  Email        string
  PasswordHash string
  CreatedAt    time.Time
}

func (s *Store) CreateUser(ctx context.Context, email string, passwordHash string) (User, error) {
  var user User
  err := s.DB.QueryRowContext(
    ctx,
    "insert into users (email, password_hash) values ($1, $2) returning id, email, password_hash, created_at",
    email,
    passwordHash,
  ).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
  if err != nil {
    return User{}, err
  }
  return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (User, error) {
  var user User
  err := s.DB.QueryRowContext(
    ctx,
    "select id, email, password_hash, created_at from users where email = $1",
    email,
  ).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
  if errors.Is(err, sql.ErrNoRows) {
    return User{}, ErrNotFound
  }
  if err != nil {
    return User{}, err
  }
  return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (User, error) {
  var user User
  err := s.DB.QueryRowContext(
    ctx,
    "select id, email, password_hash, created_at from users where id = $1",
    id,
  ).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
  if errors.Is(err, sql.ErrNoRows) {
    return User{}, ErrNotFound
  }
  if err != nil {
    return User{}, err
  }
  return user, nil
}
