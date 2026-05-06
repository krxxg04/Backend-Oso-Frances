package http

import (
	"backend-of/internal/auth"
	"backend-of/internal/config"
	"backend-of/internal/repository/jsondb"
	"backend-of/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestVehicleAndSimulationHTTPFlow(t *testing.T) {
	store, err := jsondb.NewStore(filepath.Join(t.TempDir(), "data.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	cfg := config.Config{
		JWTSecret:         "test-secret",
		AccessTTL:         15 * time.Minute,
		RefreshTTL:        time.Hour,
		LoginMaxPerMinute: 10,
	}
	tokenManager := auth.NewTokenManager(cfg.JWTSecret)
	authSvc := service.NewAuthService(store, store, tokenManager, cfg.AccessTTL, cfg.RefreshTTL)
	if err := authSvc.Seed(context.Background()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	simSvc := service.NewSimulationService(store, store)
	vehicleSvc := service.NewVehicleService(store, store)
	router := NewRouter(cfg, authSvc, simSvc, vehicleSvc)

	cookies := registerAndCookies(t, router)

	vehicleBody := `{"marca":"Toyota","modelo":"Yaris","anio":2025,"tipo":"sedan","precio":80000,"moneda":"PEN"}`
	vehicleResp := performJSON(router, http.MethodPost, "/api/v1/vehiculos", vehicleBody, cookies)
	if vehicleResp.Code != http.StatusCreated {
		t.Fatalf("expected vehicle created, got %d: %s", vehicleResp.Code, vehicleResp.Body.String())
	}

	listVehicles := performJSON(router, http.MethodGet, "/api/v1/vehiculos", "", cookies)
	if listVehicles.Code != http.StatusOK {
		t.Fatalf("expected vehicles list ok, got %d", listVehicles.Code)
	}

	simulationBody := `{
		"moneda":"PEN",
		"vehiculo":{"marca":"Toyota","modelo":"Yaris","anio":2025,"precio":80000},
		"porcentajeCuotaInicial":20,
		"plazoMeses":36,
		"tipoTasa":"nominal",
		"tasaAnual":18,
		"frecuenciaCapitalizacion":12,
		"periodosPorAnio":12,
		"cuotaFinalBalloon":24000,
		"seguroVehicularMensual":180,
		"seguroDesgravamenAnual":1.2,
		"tipoGracia":"sin_gracia",
		"fechaInicio":"2026-06-01"
	}`
	simResp := performJSON(router, http.MethodPost, "/api/v1/simulaciones", simulationBody, cookies)
	if simResp.Code != http.StatusCreated {
		t.Fatalf("expected simulation created, got %d: %s", simResp.Code, simResp.Body.String())
	}

	var sim struct {
		ID     string `json:"id"`
		Result struct {
			Resumen struct {
				TCEA float64 `json:"tcea"`
			} `json:"resumen"`
			Cronograma []struct {
				SaldoInicial float64 `json:"saldoInicial"`
				Seguro       float64 `json:"seguro"`
			} `json:"cronograma"`
		} `json:"result"`
	}
	if err := json.Unmarshal(simResp.Body.Bytes(), &sim); err != nil {
		t.Fatalf("decode sim: %v", err)
	}
	if sim.ID == "" || sim.Result.Resumen.TCEA <= 0 || len(sim.Result.Cronograma) != 36 {
		t.Fatalf("unexpected simulation response: %+v", sim)
	}
	if sim.Result.Cronograma[0].SaldoInicial <= 0 || sim.Result.Cronograma[0].Seguro <= 0 {
		t.Fatalf("expected enriched schedule row")
	}

	filtered := performJSON(router, http.MethodGet, "/api/v1/simulaciones?moneda=PEN&plazoMeses=36&vehiculo=Yaris", "", cookies)
	if filtered.Code != http.StatusOK {
		t.Fatalf("expected filtered history ok, got %d", filtered.Code)
	}

	openAPI := performJSON(router, http.MethodGet, "/openapi.json", "", nil)
	if openAPI.Code != http.StatusOK {
		t.Fatalf("expected openapi ok, got %d", openAPI.Code)
	}
}

func registerAndCookies(t *testing.T, router http.Handler) []*http.Cookie {
	t.Helper()
	body := `{"username":"cliente01","email":"cliente01@email.com","dni":"12345678","fullName":"Cliente Demo","password":"secret123"}`
	resp := performJSON(router, http.MethodPost, "/api/v1/auth/register", body, nil)
	if resp.Code != http.StatusCreated {
		t.Fatalf("register got %d: %s", resp.Code, resp.Body.String())
	}
	return resp.Result().Cookies()
}

func performJSON(router http.Handler, method, path, body string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}
