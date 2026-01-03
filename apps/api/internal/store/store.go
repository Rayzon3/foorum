package store

import (
  "database/sql"
  "errors"
)

var ErrNotFound = errors.New("not found")

type Store struct {
  Users *UserStore
}

func New(db *sql.DB) *Store {
  return &Store{
    Users: &UserStore{db: db},
  }
}
