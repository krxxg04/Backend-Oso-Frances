package domain

import "time"

type GraceType string

const (
	GraceTotal   GraceType = "total"
	GraceParcial GraceType = "parcial"
)

type Pago struct {
	Mes          int     `json:"mes"`
	Cuota        float64 `json:"cuota"`
	Interes      float64 `json:"interes"`
	Amortizacion float64 `json:"amortizacion"`
	SaldoDeudor  float64 `json:"saldoDeudor"`
}

type SimulacionInput struct {
	NombreCliente          string    `json:"nombreCliente"`
	PrecioVehiculo         float64   `json:"precioVehiculo"`
	PorcentajeCuotaInicial float64   `json:"porcentajeCuotaInicial"`
	PlazoMeses             int       `json:"plazoMeses"`
	TasaEfectivaAnual      float64   `json:"tasaEfectivaAnual"`
	PeriodosPorAnio        int       `json:"periodosPorAnio"`
	PeriodosGracia         int       `json:"periodosGracia"`
	TipoGracia             GraceType `json:"tipoGracia"`
	ValorFinal             float64   `json:"valorFinal"`
	CostosFinanciados      float64   `json:"costosFinanciados"`
	CostosIniciales        float64   `json:"costosIniciales"`
}

type SimulacionResult struct {
	TasaPeriodo     float64 `json:"tasaPeriodo"`
	CuotaBase       float64 `json:"cuotaBase"`
	VAN             float64 `json:"van"`
	TIR             float64 `json:"tir"`
	TCEA            float64 `json:"tcea"`
	CostoTotal      float64 `json:"costoTotal"`
	MontoNeto       float64 `json:"montoNeto"`
	MontoFinanciado float64 `json:"montoFinanciado"`
	Cronograma      []Pago  `json:"cronograma"`
}

type Simulacion struct {
	ID        string           `json:"id"`
	ClienteID string           `json:"clienteId"`
	CreadoEn  time.Time        `json:"creadoEn"`
	Input     SimulacionInput  `json:"input"`
	Result    SimulacionResult `json:"result"`
}

type Cliente struct {
	ID        string    `json:"id"`
	Nombre    string    `json:"nombre"`
	NombreKey string    `json:"nombreKey"`
	CreadoEn  time.Time `json:"creadoEn"`
}

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}
