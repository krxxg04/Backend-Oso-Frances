package domain

import "time"

type GraceType string

const (
	GraceNone    GraceType = "sin_gracia"
	GraceTotal   GraceType = "total"
	GraceParcial GraceType = "parcial"
)

type Currency string

const (
	CurrencyPEN Currency = "PEN"
	CurrencyUSD Currency = "USD"
)

type Vehicle struct {
	ID        string    `json:"id,omitempty"`
	ClienteID string    `json:"clienteId,omitempty"`
	Marca     string    `json:"marca,omitempty"`
	Modelo    string    `json:"modelo,omitempty"`
	Anio      int       `json:"anio,omitempty"`
	Tipo      string    `json:"tipo,omitempty"`
	Precio    float64   `json:"precio,omitempty"`
	Moneda    Currency  `json:"moneda,omitempty"`
	CreadoEn  time.Time `json:"creadoEn,omitempty"`
}

type Rate struct {
	TasaEfectivaAnual  float64 `json:"tasaEfectivaAnual,omitempty"`
	PeriodosPagoPorAnio int    `json:"periodosPagoPorAnio,omitempty"`
}

type Insurance struct {
	SeguroVehicularMensual float64 `json:"seguroVehicularMensual,omitempty"`
	SeguroDesgravamenAnual float64 `json:"seguroDesgravamenAnual,omitempty"`
}

type BankOption struct {
	ID                        string   `json:"id"`
	Nombre                    string   `json:"nombre"`
	Producto                  string   `json:"producto"`
	Moneda                    Currency `json:"moneda"`
	TasaEfectivaAnual         float64  `json:"tasaEfectivaAnual"`
	SeguroDesgravamenMensual  float64  `json:"seguroDesgravamenMensual"`
	SeguroDesgravamenAnual    float64  `json:"seguroDesgravamenAnual"`
	SeguroVehicularMensualPct float64  `json:"seguroVehicularMensualPorcentaje,omitempty"`
	MontoMin                  float64  `json:"montoMin,omitempty"`
	MontoMax                  float64  `json:"montoMax,omitempty"`
	PorcentajeCuotaInicialMin float64  `json:"porcentajeCuotaInicialMin,omitempty"`
	PorcentajeCuotaInicialMax float64  `json:"porcentajeCuotaInicialMax,omitempty"`
	PlazosMeses               []int    `json:"plazosMeses,omitempty"`
	PeriodosGraciaMax         int      `json:"periodosGraciaMax,omitempty"`
	Fuente                    string   `json:"fuente,omitempty"`
	FuenteURL                 string   `json:"fuenteUrl,omitempty"`
	Notas                     string   `json:"notas,omitempty"`
}

type Pago struct {
	Mes                 int       `json:"mes"`
	Periodo             int       `json:"periodo"`
	Fecha               string    `json:"fecha,omitempty"`
	FechaPago           time.Time `json:"fechaPago,omitempty"`
	SaldoInicial        float64   `json:"saldoInicial"`
	Cuota               float64   `json:"cuota"`
	CuotaCapitalInteres float64   `json:"cuotaCapitalInteres"`
	Interes             float64   `json:"interes"`
	SeguroVehicular     float64   `json:"seguroVehicular"`
	SeguroDesgravamen   float64   `json:"seguroDesgravamen"`
	Seguro              float64   `json:"seguro"`
	Amortizacion        float64   `json:"amortizacion"`
	SaldoFinal          float64   `json:"saldoFinal"`
	SaldoDeudor         float64   `json:"saldoDeudor"`
	TipoGracia          GraceType `json:"tipoGracia,omitempty"`
}

type SimulacionInput struct {
	NombreCliente            string    `json:"nombreCliente"`
	BancoID                  string    `json:"bancoId,omitempty"`
	Moneda                   Currency  `json:"moneda,omitempty"`
	Vehiculo                 Vehicle   `json:"vehiculo,omitempty"`
	FechaInicio              string    `json:"fechaInicio,omitempty"`
	PrecioVehiculo           float64   `json:"precioVehiculo"`
	PorcentajeCuotaInicial   float64   `json:"porcentajeCuotaInicial"`
	PlazoMeses               int       `json:"plazoMeses"`
	TasaAnual                float64   `json:"tasaAnual,omitempty"`
	TasaEfectivaAnual        float64   `json:"tasaEfectivaAnual"`
	PeriodosPorAnio          int       `json:"periodosPorAnio"`
	PeriodosGracia           int       `json:"periodosGracia"`
	TipoGracia               GraceType `json:"tipoGracia"`
	ValorFinal               float64   `json:"valorFinal"`
	CuotaFinalBalloon        float64   `json:"cuotaFinalBalloon,omitempty"`
	SeguroVehicularMensual   float64   `json:"seguroVehicularMensual,omitempty"`
	SeguroDesgravamenAnual   float64   `json:"seguroDesgravamenAnual,omitempty"`
	CostosFinanciados        float64   `json:"costosFinanciados"`
	CostosIniciales          float64   `json:"costosIniciales"`
}

type FinancialSummary struct {
	MontoFinanciado   float64 `json:"montoFinanciado"`
	CuotaInicial      float64 `json:"cuotaInicial"`
	CuotaMensual      float64 `json:"cuotaMensual"`
	CuotaFinalBalloon float64 `json:"cuotaFinalBalloon"`
	TotalIntereses    float64 `json:"totalIntereses"`
	TotalSeguros      float64 `json:"totalSeguros"`
	TotalPagado       float64 `json:"totalPagado"`
	VAN               float64 `json:"van"`
	TIR               float64 `json:"tir"`
	TCEA              float64 `json:"tcea"`
	FechaFinalizacion string  `json:"fechaFinalizacion,omitempty"`
}

type SimulacionResult struct {
	Banco           *BankOption      `json:"banco,omitempty"`
	TasaPeriodo     float64          `json:"tasaPeriodo"`
	Tasa            Rate             `json:"tasa"`
	Seguros         Insurance        `json:"seguros"`
	CuotaBase       float64          `json:"cuotaBase"`
	VAN             float64          `json:"van"`
	TIR             float64          `json:"tir"`
	TCEA            float64          `json:"tcea"`
	CostoTotal      float64          `json:"costoTotal"`
	TotalIntereses  float64          `json:"totalIntereses"`
	TotalSeguros    float64          `json:"totalSeguros"`
	CuotaInicial    float64          `json:"cuotaInicial"`
	MontoNeto       float64          `json:"montoNeto"`
	MontoFinanciado float64          `json:"montoFinanciado"`
	Resumen         FinancialSummary `json:"resumen"`
	Cronograma      []Pago           `json:"cronograma"`
}

type Simulacion struct {
	ID        string           `json:"id"`
	ClienteID string           `json:"clienteId"`
	CreadoEn  time.Time        `json:"creadoEn"`
	Input     SimulacionInput  `json:"input"`
	Result    SimulacionResult `json:"result"`
}

type SimulacionFilter struct {
	FechaDesde string   `json:"fechaDesde,omitempty"`
	FechaHasta string   `json:"fechaHasta,omitempty"`
	Moneda     Currency `json:"moneda,omitempty"`
	PlazoMeses int      `json:"plazoMeses,omitempty"`
	MontoMin   float64  `json:"montoMin,omitempty"`
	MontoMax   float64  `json:"montoMax,omitempty"`
	Vehiculo   string   `json:"vehiculo,omitempty"`
}

type Cliente struct {
	ID        string    `json:"id"`
	Nombre    string    `json:"nombre"`
	NombreKey string    `json:"nombreKey"`
	DNI       string    `json:"dni,omitempty"`
	Email     string    `json:"email,omitempty"`
	Telefono  string    `json:"telefono,omitempty"`
	Username  string    `json:"username,omitempty"`
	CreadoEn  time.Time `json:"creadoEn"`
}

type User struct {
	Username     string `json:"username"`
	Email        string `json:"email,omitempty"`
	DNI          string `json:"dni,omitempty"`
	FullName     string `json:"fullName,omitempty"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

