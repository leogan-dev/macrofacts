package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      *Repo
	jwtSecret []byte
}

func NewService(pg *pgxpool.Pool, jwtSecret []byte) *Service {
	return &Service{
		repo:      NewRepo(pg),
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(username, password string) error {
	u, err := CanonicalizeUsername(username)
	if err != nil {
		return ErrUsernameInvalid
	}
	if len(password) < 8 || len(password) > 128 {
		return ErrPasswordInvalid
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	_, err = s.repo.CreateUser(context.Background(), u, string(pwHash))
	return err
}

func (s *Service) Login(username, password string) (string, error) {
	u, err := CanonicalizeUsername(username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	user, err := s.repo.GetUserByUsername(context.Background(), u)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return IssueToken(s.jwtSecret, user.ID, user.Username, time.Now().Add(7*24*time.Hour))
}

func (s *Service) ParseToken(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("missing token")
	}
	return ParseToken(s.jwtSecret, raw)
}

func (s *Service) GetSettings(userID string) (MeSettingsResponse, error) {
	return s.repo.GetSettings(context.Background(), userID)
}

func (s *Service) UpdateSettings(userID string, req UpdateSettingsRequest) (MeSettingsResponse, error) {
	if err := s.repo.UpsertSettings(context.Background(), userID, req); err != nil {
		return MeSettingsResponse{}, err
	}
	return s.repo.GetSettings(context.Background(), userID)
}
