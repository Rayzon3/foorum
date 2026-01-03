package store

import (
	"context"
	"database/sql"
	"errors"

	"jabber_v3/apps/api/internal/models"
	"jabber_v3/apps/api/internal/store/qb"
)

type UserStore struct {
  db *sql.DB
}

var userColumns = []string{"id", "email", "username", "password_hash", "created_at"}

type rowScanner interface {
  Scan(dest ...any) error
}

func scanUser(row rowScanner) (models.User, error) {
  var user models.User
  err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}

func (s *UserStore) CreateUser(ctx context.Context, email string, username string, passwordHash string) (models.User, error) {
  query, args := qb.Insert("users").
    Columns("email", "username", "password_hash").
    Values(email, username, passwordHash).
    Returning(userColumns...).
    Build()
  row := s.db.QueryRowContext(ctx, query, args...)
  user, err := scanUser(row)
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
  query, args := qb.Select(userColumns...).
    From("users").
    WhereEq("email", email).
    Build()
  row := s.db.QueryRowContext(ctx, query, args...)
  user, err := scanUser(row)
  if errors.Is(err, sql.ErrNoRows) {
    return models.User{}, ErrNotFound
  }
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}

func (s *UserStore) GetUserByID(ctx context.Context, id string) (models.User, error) {
  query, args := qb.Select(userColumns...).
    From("users").
    WhereEq("id", id).
    Build()
  row := s.db.QueryRowContext(ctx, query, args...)
  user, err := scanUser(row)
  if errors.Is(err, sql.ErrNoRows) {
    return models.User{}, ErrNotFound
  }
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}

func (s *UserStore) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
  query, args := qb.Select(userColumns...).
    From("users").
    WhereEq("username", username).
    Build()
  row := s.db.QueryRowContext(ctx, query, args...)
  user, err := scanUser(row)
  if errors.Is(err, sql.ErrNoRows) {
    return models.User{}, ErrNotFound
  }
  if err != nil {
    return models.User{}, err
  }
  return user, nil
}
