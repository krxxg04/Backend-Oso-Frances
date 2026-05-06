package service

import (
	"backend-of/internal/domain"
	"math"
	"testing"
)

func TestCalculateSimulation(t *testing.T) {
	in := domain.SimulacionInput{
		NombreCliente: "Juan Perez", PrecioVehiculo: 50000, PorcentajeCuotaInicial: 10,
		PlazoMeses: 24, TasaEfectivaAnual: 12, PeriodosPorAnio: 12, PeriodosGracia: 2, TipoGracia: domain.GraceParcial,
	}
	res := CalculateSimulation(in)
	if len(res.Cronograma) != 24 {
		t.Fatalf("expected 24 payments, got %d", len(res.Cronograma))
	}
	last := res.Cronograma[len(res.Cronograma)-1]
	if math.Abs(last.SaldoDeudor) > 0.05 {
		t.Fatalf("expected ending balance near 0, got %.4f", last.SaldoDeudor)
	}
}

func TestCalculateSimulationWithVehicleInsuranceAndSummary(t *testing.T) {
	in := domain.SimulacionInput{
		NombreCliente:            "Ana Lopez",
		Moneda:                   domain.CurrencyPEN,
		Vehiculo:                 domain.Vehicle{Marca: "Toyota", Modelo: "Yaris", Anio: 2025, Precio: 80000},
		PorcentajeCuotaInicial:   20,
		PlazoMeses:               36,
		TipoTasa:                 domain.RateNominal,
		TasaAnual:                18,
		FrecuenciaCapitalizacion: 12,
		PeriodosPorAnio:          12,
		CuotaFinalBalloon:        24000,
		SeguroVehicularMensual:   180,
		SeguroDesgravamenAnual:   1.2,
		FechaInicio:              "2026-06-01",
	}

	res := CalculateSimulation(in)

	if len(res.Cronograma) != 36 {
		t.Fatalf("expected 36 payments, got %d", len(res.Cronograma))
	}
	if res.MontoFinanciado != 64000 {
		t.Fatalf("expected financed amount 64000, got %.2f", res.MontoFinanciado)
	}
	if res.Resumen.CuotaInicial != 16000 {
		t.Fatalf("expected down payment 16000, got %.2f", res.Resumen.CuotaInicial)
	}
	if res.Resumen.CuotaFinalBalloon != 24000 {
		t.Fatalf("expected balloon 24000, got %.2f", res.Resumen.CuotaFinalBalloon)
	}
	if res.TotalSeguros <= 0 {
		t.Fatalf("expected insurance total > 0")
	}
	if res.Tasa.Tipo != domain.RateNominal {
		t.Fatalf("expected nominal rate metadata")
	}
	first := res.Cronograma[0]
	if first.Fecha != "2026-07-01" {
		t.Fatalf("expected first payment date 2026-07-01, got %s", first.Fecha)
	}
	if first.SaldoInicial != 64000 || first.Seguro <= 0 || first.SaldoFinal <= 0 {
		t.Fatalf("expected enriched payment fields, got %+v", first)
	}
}

func TestValidateSimulationInputCompraInteligente(t *testing.T) {
	in := domain.SimulacionInput{
		NombreCliente:          "Cliente",
		PrecioVehiculo:         60000,
		PorcentajeCuotaInicial: 10,
		PlazoMeses:             30,
		TasaAnual:              12,
	}

	errs := ValidateSimulationInput(in)
	if len(errs) == 0 {
		t.Fatalf("expected validation error for unsupported term")
	}
}
