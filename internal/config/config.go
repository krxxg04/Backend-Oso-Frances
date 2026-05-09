package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	JWTSecret         string
	DataFilePath      string
	DatabaseURL       string
	FrontendOrigins   string
	CookieSameSite    string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	CookieSecure      bool
	LoginMaxPerMinute int
}

func Load() Config {
	return Config{
		Port:              getenv("PORT", "8080"),
		JWTSecret:         getenv("JWT_SECRET", "change-this-secret"),
		DataFilePath:      getenv("DATA_FILE", "./data.json"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		FrontendOrigins:   getenv("FRONTEND_ORIGINS", "https://oso-frances.vercel.app,https://www.oso-frances.vercel.app,http://localhost:4321,http://localhost:4321/"),
		CookieSameSite:    getenv("COOKIE_SAMESITE", "none"),
		AccessTTL:         time.Duration(getenvInt("ACCESS_TTL_MIN", 15)) * time.Minute,
		RefreshTTL:        time.Duration(getenvInt("REFRESH_TTL_HOURS", 24)) * time.Hour,
		CookieSecure:      getenv("COOKIE_SECURE", "true") == "true",
		LoginMaxPerMinute: getenvInt("LOGIN_MAX_PER_MIN", 10),
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
