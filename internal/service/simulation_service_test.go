package service

import (
	"backend-of/internal/domain"
	"math"
	"testing"
)

func TestCalculateSimulation(t *testing.T) {
	in := domain.SimulacionInput{
		NombreCliente: "Juan Perez", PrecioVehiculo: 50000, PorcentajeCuotaInicial: 10,
		PlazoMeses: 24, TasaNominalAnual: 12, CapitalizacionPorAnio: 12, PeriodosGracia: 2, TipoGracia: domain.GraceParcial,
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
