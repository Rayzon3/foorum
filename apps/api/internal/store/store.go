package store

import (
  "database/sql"
  "errors"
)

var ErrNotFound = errors.New("not found")

type Store struct {
  Users *UserStore
  Posts *PostStore
  Votes *VoteStore
}

func New(db *sql.DB) *Store {
  return &Store{
    Users: &UserStore{db: db},
    Posts: &PostStore{db: db},
    Votes: &VoteStore{db: db},
  }
}
