package service

import (
	"backend-of/internal/auth"
	"backend-of/internal/domain"
	"backend-of/internal/repository"
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users      repository.UserRepository
	clientes   repository.ClienteRepository
	tokens     *auth.TokenManager
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(users repository.UserRepository, clientes repository.ClienteRepository, tokens *auth.TokenManager, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{users: users, clientes: clientes, tokens: tokens, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (s *AuthService) Seed(ctx context.Context) error {
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	userHash, _ := bcrypt.GenerateFromPassword([]byte("user"), bcrypt.DefaultCost)
	if err := s.users.SeedIfEmpty(ctx, []domain.User{
		{Username: "admin", PasswordHash: string(adminHash), Role: "admin"},
		{Username: "user", PasswordHash: string(userHash), Role: "user"},
	}); err != nil {
		return err
	}
	_, _ = s.clientes.GetOrCreateByNombre(ctx, "admin")
	_, _ = s.clientes.GetOrCreateByNombre(ctx, "user")
	return nil
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

func (s *AuthService) GetUser(ctx context.Context, username string) (domain.User, bool, error) {
	return s.users.GetByUsername(ctx, username)
}

func (s *AuthService) Register(ctx context.Context, username, password string) (string, string, error) {
	return s.RegisterProfile(ctx, domain.User{Username: username}, password)
}

func (s *AuthService) RegisterProfile(ctx context.Context, user domain.User, password string) (string, string, error) {
	username := strings.TrimSpace(user.Username)
	user.Username = username
	user.Email = strings.TrimSpace(user.Email)
	user.DNI = strings.TrimSpace(user.DNI)
	user.FullName = strings.TrimSpace(user.FullName)
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(password) < 6 {
		return "", "", errors.New("validation_error")
	}
	if strings.Contains(username, " ") {
		return "", "", errors.New("validation_error")
	}
	if user.Email != "" && !strings.Contains(user.Email, "@") {
		return "", "", errors.New("validation_error")
	}
	if user.DNI != "" && len(user.DNI) != 8 {
		return "", "", errors.New("validation_error")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	user.PasswordHash = string(hash)
	user.Role = "user"
	if err := s.users.CreateUser(ctx, user); err != nil {
		if err.Error() == "user_exists" {
			return "", "", errors.New("conflict")
		}
		return "", "", err
	}
	clientName := user.FullName
	if clientName == "" {
		clientName = username
	}
	if _, err := s.clientes.GetOrCreateByNombre(ctx, clientName); err != nil {
		return "", "", err
	}
	acc, err := s.tokens.Sign(username, "access", time.Now().Add(s.accessTTL))
	if err != nil {
		return "", "", err
	}
	ref, err := s.tokens.Sign(username, "refresh", time.Now().Add(s.refreshTTL))
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
