package rest

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const googleStateCookie = "google_oauth_state"

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (h *Handler) GoogleLogin(c *gin.Context) {
	if h.cfg.GoogleClientID == "" || h.cfg.GoogleRedirectURI == "" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "google oauth no configurado"})
		return
	}
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo iniciar google oauth"})
		return
	}
	c.SetSameSite(cookieSameSiteMode(h.cfg.CookieSameSite))
	c.SetCookie(googleStateCookie, state, 600, "/", "", h.cfg.CookieSecure, true)
	c.Redirect(http.StatusFound, googleAuthURL(h.cfg.GoogleClientID, h.cfg.GoogleRedirectURI, h.cfg.GoogleScopes, state))
}

func (h *Handler) GoogleCallback(c *gin.Context) {
	if h.cfg.GoogleClientID == "" || h.cfg.GoogleClientSecret == "" || h.cfg.GoogleRedirectURI == "" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "google oauth no configurado"})
		return
	}
	state := c.Query("state")
	code := c.Query("code")
	expectedState, err := c.Cookie(googleStateCookie)
	if err != nil || expectedState == "" || state == "" || expectedState != state {
		h.redirectOAuthError(c, "state_invalido")
		return
	}
	token, err := exchangeGoogleCode(c.Request.Context(), h.cfg.GoogleClientID, h.cfg.GoogleClientSecret, h.cfg.GoogleRedirectURI, code)
	if err != nil {
		h.redirectOAuthError(c, "token_exchange_error")
		return
	}
	profile, err := fetchGoogleUserInfo(c.Request.Context(), token.AccessToken)
	if err != nil || !profile.EmailVerified {
		h.redirectOAuthError(c, "google_profile_error")
		return
	}
	acc, ref, err := h.authSvc.LoginWithGoogle(c, profile.Sub, profile.Email, profile.Name, profile.Picture)
	if err != nil {
		h.redirectOAuthError(c, "google_login_error")
		return
	}
	h.setAuthCookies(c, acc, ref)
	h.clearGoogleStateCookie(c)
	c.Redirect(http.StatusFound, h.frontendRedirectURL("auth=success&provider=google"))
}

func (h *Handler) redirectOAuthError(c *gin.Context, reason string) {
	h.clearGoogleStateCookie(c)
	c.Redirect(http.StatusFound, h.frontendRedirectURL("auth=error&provider=google&reason="+url.QueryEscape(reason)))
}

func (h *Handler) clearGoogleStateCookie(c *gin.Context) {
	c.SetSameSite(cookieSameSiteMode(h.cfg.CookieSameSite))
	c.SetCookie(googleStateCookie, "", -1, "/", "", h.cfg.CookieSecure, true)
}

func (h *Handler) frontendRedirectURL(rawQuery string) string {
	origins := strings.Split(h.cfg.FrontendOrigins, ",")
	target := "http://localhost:4321"
	if len(origins) > 0 && strings.TrimSpace(origins[0]) != "" {
		target = strings.TrimRight(strings.TrimSpace(origins[0]), "/")
	}
	if rawQuery == "" {
		return target
	}
	return target + "/?" + rawQuery
}

func randomState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func googleAuthURL(clientID, redirectURI string, scopes []string, state string) string {
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("response_type", "code")
	values.Set("scope", strings.Join(scopes, " "))
	values.Set("state", state)
	values.Set("access_type", "offline")
	values.Set("prompt", "consent")
	return "https://accounts.google.com/o/oauth2/v2/auth?" + values.Encode()
}

func exchangeGoogleCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (googleTokenResponse, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return googleTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return googleTokenResponse{}, err
	}
	defer resp.Body.Close()

	var out googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return googleTokenResponse{}, err
	}
	if resp.StatusCode >= 400 || out.AccessToken == "" {
		return googleTokenResponse{}, errors.New("google_token_exchange_failed")
	}
	return out, nil
}

func fetchGoogleUserInfo(ctx context.Context, accessToken string) (googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return googleUserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return googleUserInfo{}, err
	}
	defer resp.Body.Close()

	var out googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return googleUserInfo{}, err
	}
	if resp.StatusCode >= 400 || out.Sub == "" || out.Email == "" {
		return googleUserInfo{}, errors.New("google_userinfo_failed")
	}
	return out, nil
}
