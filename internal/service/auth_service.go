package service

import (
	"backend-of/internal/auth"
	"backend-of/internal/domain"
	"backend-of/internal/repository"
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users      repository.UserRepository
	tokens     *auth.TokenManager
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(users repository.UserRepository, tokens *auth.TokenManager, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{users: users, tokens: tokens, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (s *AuthService) Seed(ctx context.Context) error {
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	userHash, _ := bcrypt.GenerateFromPassword([]byte("user"), bcrypt.DefaultCost)
	return s.users.SeedIfEmpty(ctx, []domain.User{
		{Username: "admin", PasswordHash: string(adminHash), Role: "admin"},
		{Username: "user", PasswordHash: string(userHash), Role: "user"},
	})
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, error) {
	u, ok, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", errors.New("unauthorized")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", errors.New("unauthorized")
	}
	acc, err := s.tokens.Sign(u.Username, "access", time.Now().Add(s.accessTTL))
	if err != nil {
		return "", "", err
	}
	ref, err := s.tokens.Sign(u.Username, "refresh", time.Now().Add(s.refreshTTL))
	if err != nil {
		return "", "", err
	}
	return acc, ref, nil
}

func (s *AuthService) VerifyAccess(token string) (string, error) {
	claims, err := s.tokens.Verify(token, "access")
	if err != nil {
		return "", err
	}
	return claims.Sub, nil
}

func (s *AuthService) Refresh(refresh string) (string, string, error) {
	claims, err := s.tokens.Verify(refresh, "refresh")
	if err != nil {
		return "", "", errors.New("unauthorized")
	}
	acc, err := s.tokens.Sign(claims.Sub, "access", time.Now().Add(s.accessTTL))
	if err != nil {
		return "", "", err
	}
	ref, err := s.tokens.Sign(claims.Sub, "refresh", time.Now().Add(s.refreshTTL))
	if err != nil {
		return "", "", err
	}
	return acc, ref, nil
}
