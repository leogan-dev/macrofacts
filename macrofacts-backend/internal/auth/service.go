package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db        *pgxpool.Pool
	jwtSecret []byte
}

func NewService(db *pgxpool.Pool, jwtSecret []byte) *Service {
	return &Service{db: db, jwtSecret: jwtSecret}
}

func (s *Service) Register(username, password string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 32 {
		return errors.New("username must be 3-32 chars")
	}
	if len(password) < 8 || len(password) > 128 {
		return errors.New("password must be 8-128 chars")
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	_, err = s.db.Exec(
		contextOrBackground(),
		`insert into users (username, password_hash) values ($1, $2)`,
		username, string(hashBytes),
	)
	if err != nil {
		return errors.New("username already exists")
	}
	return nil
}

func (s *Service) Login(username, password string) (token string, userID string, cleanUsername string, err error) {
	username = strings.TrimSpace(username)

	var id string
	var un string
	var hash string

	e := s.db.QueryRow(contextOrBackground(),
		`select id::text, username, password_hash from users where username = $1`,
		username,
	).Scan(&id, &un, &hash)

	if e != nil {
		_ = bcrypt.CompareHashAndPassword([]byte(fakeBcryptHash()), []byte(password))
		return "", "", "", errors.New("invalid credentials")
	}

	if e := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); e != nil {
		return "", "", "", errors.New("invalid credentials")
	}

	tok, e := s.makeJWT(id, un)
	if e != nil {
		return "", "", "", errors.New("failed to sign token")
	}

	return tok, id, un, nil
}

func (s *Service) makeJWT(userID, username string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"iat":      now.Unix(),
		"exp":      now.Add(14 * 24 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.jwtSecret)
}

func (s *Service) ParseToken(raw string) (userID string, username string, err error) {
	claims := jwt.MapClaims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return "", "", errors.New("invalid token")
	}

	uid, _ := claims["sub"].(string)
	un, _ := claims["username"].(string)

	if subtle.ConstantTimeEq(int32(len(uid)), 0) == 1 || subtle.ConstantTimeEq(int32(len(un)), 0) == 1 {
		return "", "", errors.New("invalid token")
	}

	return uid, un, nil
}

// keep behavior stable even without request ctx for now
func contextOrBackground() context.Context {
	return context.Background()
}

func fakeBcryptHash() string {
	return "$2a$10$yAq4qfJc6Q7xjYHcqJq3gO9YorQWQm/1cSx4xV6qY9tUqQf2mFq1S"
}
