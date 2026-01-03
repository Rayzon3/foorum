package types

import "time"

type AuthUser struct {
  ID    string
  Email string
}

type CredentialsRequest struct {
  Email    string `json:"email"`
  Username string `json:"username"`
  Password string `json:"password"`
}

type AuthResponse struct {
  Token string   `json:"token"`
  User  UserView `json:"user"`
}

type UserView struct {
  ID        string    `json:"id"`
  Email     string    `json:"email"`
  Username  string    `json:"username"`
  CreatedAt time.Time `json:"createdAt"`
}
