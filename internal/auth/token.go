package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Claims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
	Typ string `json:"typ"`
}

type TokenManager struct {
	secret []byte
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret)}
}

func (t *TokenManager) Sign(sub, typ string, exp time.Time) (string, error) {
	c := Claims{Sub: sub, Exp: exp.Unix(), Typ: typ}
	payload, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	sig := t.sign(encoded)
	return encoded + "." + sig, nil
}

func (t *TokenManager) Verify(token, expectedType string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, errors.New("invalid token format")
	}
	if !hmac.Equal([]byte(t.sign(parts[0])), []byte(parts[1])) {
		return Claims{}, errors.New("invalid token signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, err
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return Claims{}, err
	}
	if claims.Typ != expectedType {
		return Claims{}, errors.New("invalid token type")
	}
	if time.Now().Unix() > claims.Exp {
		return Claims{}, errors.New("token expired")
	}
	return claims, nil
}

func (t *TokenManager) sign(payload string) string {
	h := hmac.New(sha256.New, t.secret)
	h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
