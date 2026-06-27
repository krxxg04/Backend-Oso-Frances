package services

import (
	domain "backend-of/internal/domain/entities"
	"backend-of/internal/domain/repositories"
	"backend-of/internal/infrastructure/security"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users      repository.UserRepository
	tokens     *security.TokenManager
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(users repository.UserRepository, tokens *security.TokenManager, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{users: users, tokens: tokens, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (s *AuthService) Seed(ctx context.Context) error {
	_ = ctx
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

func (s *AuthService) UpdateProfile(ctx context.Context, username string, update domain.UserProfileUpdate) (domain.User, bool, error) {
	update.Email = strings.TrimSpace(update.Email)
	update.DNI = strings.TrimSpace(update.DNI)
	update.FullName = strings.TrimSpace(update.FullName)
	update.PictureURL = strings.TrimSpace(update.PictureURL)

	if update.Email != "" && !strings.Contains(update.Email, "@") {
		return domain.User{}, false, errors.New("validation_error")
	}
	if update.DNI != "" && len(update.DNI) != 8 {
		return domain.User{}, false, errors.New("validation_error")
	}
	if update.PictureURL != "" && !isPNGPicture(update.PictureURL) {
		return domain.User{}, false, errors.New("validation_error")
	}
	return s.users.UpdateProfileByUsername(ctx, username, update)
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
		if err.Error() == "email_exists" {
			return "", "", errors.New("email_conflict")
		}
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

func (s *AuthService) RefreshWithSubject(refresh string) (string, string, string, error) {
	claims, err := s.tokens.Verify(refresh, "refresh")
	if err != nil {
		return "", "", "", errors.New("unauthorized")
	}
	acc, err := s.tokens.Sign(claims.Sub, "access", time.Now().Add(s.accessTTL))
	if err != nil {
		return "", "", "", err
	}
	ref, err := s.tokens.Sign(claims.Sub, "refresh", time.Now().Add(s.refreshTTL))
	if err != nil {
		return "", "", "", err
	}
	return claims.Sub, acc, ref, nil
}

func (s *AuthService) LoginWithGoogle(ctx context.Context, googleID, email, fullName, pictureURL string) (string, string, error) {
	googleID = strings.TrimSpace(googleID)
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)
	pictureURL = strings.TrimSpace(pictureURL)

	if googleID == "" || email == "" {
		return "", "", errors.New("validation_error")
	}

	user, ok, err := s.users.GetByGoogleID(ctx, googleID)
	if err != nil {
		return "", "", err
	}
	if ok {
		user, _, err = s.users.LinkGoogleAccount(ctx, user.Username, googleID, email, fullName, pictureURL)
		if err != nil {
			return "", "", err
		}
		return s.issueSessionTokens(user.Username)
	}

	user, ok, err = s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if ok {
		user, _, err = s.users.LinkGoogleAccount(ctx, user.Username, googleID, email, fullName, pictureURL)
		if err != nil {
			return "", "", err
		}
		return s.issueSessionTokens(user.Username)
	}

	username, err := s.generateAvailableUsername(ctx, email, fullName)
	if err != nil {
		return "", "", err
	}
	newUser := domain.User{
		Username:   username,
		Email:      email,
		FullName:   fullName,
		GoogleID:   googleID,
		PictureURL: pictureURL,
		Role:       "user",
	}
	if err := s.users.CreateUser(ctx, newUser); err != nil {
		if err.Error() == "user_exists" {
			return "", "", errors.New("conflict")
		}
		return "", "", err
	}
	return s.issueSessionTokens(username)
}

func isPNGPicture(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.HasSuffix(normalized, ".png") || strings.HasPrefix(normalized, "data:image/png;base64,")
}

func (s *AuthService) issueSessionTokens(username string) (string, string, error) {
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

var usernameSanitizer = regexp.MustCompile(`[^a-z0-9_]+`)

func (s *AuthService) generateAvailableUsername(ctx context.Context, email, fullName string) (string, error) {
	base := usernameBaseFromIdentity(email, fullName)
	for i := 0; i < 100; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s%d", base, i+1)
		}
		if len(candidate) < 3 {
			candidate = candidate + "user"
		}
		_, exists, err := s.users.GetByUsername(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("internal_error")
}

func usernameBaseFromIdentity(email, fullName string) string {
	raw := fullName
	if raw == "" {
		raw, _, _ = strings.Cut(email, "@")
	}
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.ReplaceAll(raw, " ", "_")
	raw = usernameSanitizer.ReplaceAllString(raw, "")
	raw = strings.Trim(raw, "_")
	if raw == "" {
		return "user"
	}
	if len(raw) > 20 {
		raw = raw[:20]
	}
	return raw
}
