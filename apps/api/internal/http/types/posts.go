package types

import "time"

type CreatePostRequest struct {
  Title string `json:"title"`
  Body  string `json:"body"`
}

type VoteRequest struct {
  Value int `json:"value"`
}

type PostView struct {
  ID        string    `json:"id"`
  Title     string    `json:"title"`
  Body      string    `json:"body"`
  CreatedAt time.Time `json:"createdAt"`
  Author    AuthorView `json:"author"`
  Score     int       `json:"score"`
  MyVote    int       `json:"myVote"`
}

type AuthorView struct {
  ID       string `json:"id"`
  Email    string `json:"email"`
  Username string `json:"username"`
}
