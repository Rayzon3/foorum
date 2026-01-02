package auth

import (
  "errors"
  "time"

  "github.com/golang-jwt/jwt/v5"
  "golang.org/x/crypto/bcrypt"
)

type JWTManager struct {
  Secret []byte
  TTL    time.Duration
}

type Claims struct {
  UserID string `json:"uid"`
  jwt.RegisteredClaims
}

func (m JWTManager) Generate(userID string) (string, error) {
  now := time.Now()
  claims := Claims{
    UserID: userID,
    RegisteredClaims: jwt.RegisteredClaims{
      IssuedAt:  jwt.NewNumericDate(now),
      ExpiresAt: jwt.NewNumericDate(now.Add(m.TTL)),
    },
  }
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  return token.SignedString(m.Secret)
}

func (m JWTManager) Parse(tokenString string) (Claims, error) {
  token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
    if token.Method != jwt.SigningMethodHS256 {
      return nil, errors.New("unexpected signing method")
    }
    return m.Secret, nil
  })
  if err != nil {
    return Claims{}, err
  }
  claims, ok := token.Claims.(*Claims)
  if !ok || !token.Valid {
    return Claims{}, errors.New("invalid token")
  }
  return *claims, nil
}

func HashPassword(password string) (string, error) {
  hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
  if err != nil {
    return "", err
  }
  return string(hash), nil
}

func CheckPassword(hash string, password string) error {
  return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
