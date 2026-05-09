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
		NombreCliente:          "Ana Lopez",
		Moneda:                 domain.CurrencyPEN,
		Vehiculo:               domain.Vehicle{Marca: "Toyota", Modelo: "Yaris", Anio: 2025, Precio: 80000},
		PorcentajeCuotaInicial: 20,
		PlazoMeses:             36,
		TasaEfectivaAnual:      18,
		PeriodosPorAnio:        12,
		CuotaFinalBalloon:      24000,
		SeguroVehicularMensual: 180,
		SeguroDesgravamenAnual: 1.2,
		FechaInicio:            "2026-06-01",
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
	if math.Abs(res.Tasa.TasaEfectivaAnual-0.18) > 0.001 {
		t.Fatalf("expected effective annual rate metadata")
	}
	first := res.Cronograma[0]
	if first.Fecha != "2026-07-01" {
		t.Fatalf("expected first payment date 2026-07-01, got %s", first.Fecha)
	}
	if first.SaldoInicial != 64000 || first.Seguro <= 0 || first.SaldoFinal <= 0 {
		t.Fatalf("expected enriched payment fields, got %+v", first)
	}
}

func TestCalculateSimulationWithBankOption(t *testing.T) {
	in := domain.SimulacionInput{
		NombreCliente:          "Cliente Banco",
		BancoID:                "bbva-vehicular-sostenible",
		PrecioVehiculo:         90000,
		PorcentajeCuotaInicial: 20,
		PlazoMeses:             24,
		PeriodosGracia:         1,
		TipoGracia:             domain.GraceTotal,
	}

	res := CalculateSimulation(in)

	if res.Banco == nil || res.Banco.ID != "bbva-vehicular-sostenible" {
		t.Fatalf("expected selected bank in result, got %+v", res.Banco)
	}
	if math.Abs(res.Tasa.TasaEfectivaAnual-0.1149) > 0.001 {
		t.Fatalf("expected bank effective rate, got %+v", res.Tasa)
	}
	if res.Seguros.SeguroDesgravamenAnual <= 0 {
		t.Fatalf("expected bank desgravamen annual rate")
	}
	if len(res.Cronograma) != 24 || res.Cronograma[0].TipoGracia != domain.GraceTotal {
		t.Fatalf("expected total grace in first payment")
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

func TestValidateSimulationInputGraceOptions(t *testing.T) {
	base := domain.SimulacionInput{
		NombreCliente:          "Cliente",
		PrecioVehiculo:         60000,
		PorcentajeCuotaInicial: 10,
		PlazoMeses:             24,
		TasaAnual:              12,
	}

	validGraceTypes := []domain.GraceType{domain.GraceNone, domain.GraceParcial, domain.GraceTotal}
	for _, graceType := range validGraceTypes {
		in := base
		in.TipoGracia = graceType
		if graceType != domain.GraceNone {
			in.PeriodosGracia = 1
		}
		if errs := ValidateSimulationInput(in); len(errs) > 0 {
			t.Fatalf("expected %s to be valid, got %+v", graceType, errs)
		}
	}

	invalid := base
	invalid.TipoGracia = domain.GraceNone
	invalid.PeriodosGracia = 1
	if errs := ValidateSimulationInput(invalid); len(errs) == 0 {
		t.Fatalf("expected grace periods with sin_gracia to be invalid")
	}
}
