package models

import "time"

type Post struct {
  ID        string
  UserID    string
  Title     string
  Body      string
  CreatedAt time.Time
}

type PostWithStats struct {
  Post
  AuthorEmail    string
  AuthorUsername string
  Score          int
  MyVote         int
}
