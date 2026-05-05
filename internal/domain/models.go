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
	TasaNominalAnual       float64   `json:"tasaNominalAnual"`
	CapitalizacionPorAnio  int       `json:"capitalizacionPorAnio"`
	PeriodosGracia         int       `json:"periodosGracia"`
	TipoGracia             GraceType `json:"tipoGracia"`
}

type SimulacionResult struct {
	TasaPeriodo float64 `json:"tasaPeriodo"`
	VAN         float64 `json:"van"`
	TIR         float64 `json:"tir"`
	Cronograma  []Pago  `json:"cronograma"`
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
