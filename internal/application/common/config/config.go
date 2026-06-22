package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port               string
	JWTSecret          string
	DataFilePath       string
	DatabaseURL        string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
	GoogleScopes       []string
	FrontendOrigins    string
	CookieSameSite     string
	AccessTTL          time.Duration
	RefreshTTL         time.Duration
	CookieSecure       bool
	LoginMaxPerMinute  int
}

func Load() Config {
	loadLocalEnv()
	return Config{
		Port:               getenv("PORT", "8080"),
		JWTSecret:          getenv("JWT_SECRET", "change-this-secret"),
		DataFilePath:       getenv("DATA_FILE", "./data.json"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURI:  os.Getenv("GOOGLE_REDIRECT_URI"),
		GoogleScopes:       splitCSV(getenv("GOOGLE_SCOPES", "openid,email,profile")),
		FrontendOrigins:    getenv("FRONTEND_ORIGINS", "https://oso-frances.vercel.app,https://www.oso-frances.vercel.app,http://localhost:4321,http://localhost:4321/"),
		CookieSameSite:     getenv("COOKIE_SAMESITE", "none"),
		AccessTTL:          time.Duration(getenvInt("ACCESS_TTL_MIN", 15)) * time.Minute,
		RefreshTTL:         time.Duration(getenvInt("REFRESH_TTL_HOURS", 24)) * time.Hour,
		CookieSecure:       getenv("COOKIE_SECURE", "true") == "true",
		LoginMaxPerMinute:  getenvInt("LOGIN_MAX_PER_MIN", 10),
	}
}

func loadLocalEnv() {
	for _, path := range []string{".env.local", ".env"} {
		loadEnvFile(path)
	}
}

func loadEnvFile(path string) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func getenv(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}

func getenvInt(k string, fallback int) int {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
