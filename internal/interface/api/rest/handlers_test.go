package rest

import (
	"backend-of/internal/application/common/config"
	"backend-of/internal/application/services"
	"backend-of/internal/infrastructure/db/jsondb"
	"backend-of/internal/infrastructure/security"
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
	tokenManager := security.NewTokenManager(cfg.JWTSecret)
	authSvc := services.NewAuthService(store, store, tokenManager, cfg.AccessTTL, cfg.RefreshTTL)
	if err := authSvc.Seed(context.Background()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	simSvc := services.NewSimulationService(store, store)
	vehicleSvc := services.NewVehicleService(store, store)
	router := NewRouter(cfg, authSvc, simSvc, vehicleSvc)

	cookies := registerAndCookies(t, router)

	banksResp := performJSON(router, http.MethodGet, "/api/v1/bancos", "", nil)
	if banksResp.Code != http.StatusOK {
		t.Fatalf("expected banks catalog ok, got %d", banksResp.Code)
	}

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

	bankSimulationBody := `{
		"bancoId":"bbva-vehicular-sostenible",
		"vehiculo":{"marca":"Toyota","modelo":"Yaris","anio":2025,"precio":90000},
		"porcentajeCuotaInicial":20,
		"plazoMeses":24,
		"periodosPorAnio":12,
		"periodosGracia":1,
		"tipoGracia":"total"
	}`
	bankSimResp := performJSON(router, http.MethodPost, "/api/v1/simulaciones", bankSimulationBody, cookies)
	if bankSimResp.Code != http.StatusCreated {
		t.Fatalf("expected bank simulation created, got %d: %s", bankSimResp.Code, bankSimResp.Body.String())
	}

	filtered := performJSON(router, http.MethodGet, "/api/v1/simulaciones?moneda=PEN&plazoMeses=36&vehiculo=Yaris", "", cookies)
	if filtered.Code != http.StatusOK {
		t.Fatalf("expected filtered history ok, got %d", filtered.Code)
	}

	openAPI := performJSON(router, http.MethodGet, "/openapi.json", "", nil)
	if openAPI.Code != http.StatusOK {
		t.Fatalf("expected openapi ok, got %d", openAPI.Code)
	}

	swagger := performJSON(router, http.MethodGet, "/swagger", "", nil)
	if swagger.Code != http.StatusOK {
		t.Fatalf("expected swagger ui ok, got %d", swagger.Code)
	}
	if !bytes.Contains(swagger.Body.Bytes(), []byte("Redoc.init('/openapi.json'")) {
		t.Fatalf("expected docs ui to reference openapi.json")
	}

	docs := performJSON(router, http.MethodGet, "/docs", "", nil)
	if docs.Code != http.StatusOK {
		t.Fatalf("expected docs ui ok, got %d", docs.Code)
	}
}

func registerAndCookies(t *testing.T, router http.Handler) []*http.Cookie {
	t.Helper()
	body := `{"username":"cliente01","gmail":"cliente01@email.com","dni":"12345678","password":"secret123","repeatPassword":"secret123"}`
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
