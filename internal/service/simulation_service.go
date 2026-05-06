package service

import (
	"backend-of/internal/domain"
	"backend-of/internal/repository"
	"backend-of/internal/util"
	"context"
	"errors"
	"math"
)

type SimulationService struct {
	clientes repository.ClienteRepository
	sims     repository.SimulacionRepository
}

func NewSimulationService(clientes repository.ClienteRepository, sims repository.SimulacionRepository) *SimulationService {
	return &SimulationService{clientes: clientes, sims: sims}
}

func (s *SimulationService) CreateForUser(ctx context.Context, username string, in domain.SimulacionInput) (domain.Simulacion, error) {
	in.NombreCliente = username
	if errs := ValidateSimulationInput(in); len(errs) > 0 {
		return domain.Simulacion{}, errors.New("validation_error")
	}
	res := CalculateSimulation(in)
	c, err := s.clientes.GetOrCreateByNombre(ctx, username)
	if err != nil {
		return domain.Simulacion{}, err
	}
	in.NombreCliente = c.Nombre
	return s.sims.Create(ctx, domain.Simulacion{ClienteID: c.ID, Input: in, Result: res})
}

func (s *SimulationService) ListByUser(ctx context.Context, username string) ([]domain.Simulacion, error) {
	clientes, err := s.clientes.GetByNombre(ctx, username)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Simulacion, 0)
	for _, c := range clientes {
		items, err := s.sims.ListByClienteID(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

func (s *SimulationService) GetByID(ctx context.Context, id string) (domain.Simulacion, bool, error) {
	return s.sims.GetByID(ctx, id)
}

func ValidateSimulationInput(in domain.SimulacionInput) []domain.APIError {
	errs := make([]domain.APIError, 0)
	if util.NormalizeKey(in.NombreCliente) == "" {
		errs = append(errs, domain.NewError("validation_error", "nombreCliente es requerido", "nombreCliente"))
	}
	if in.PrecioVehiculo < 0 {
		errs = append(errs, domain.NewError("validation_error", "precioVehiculo debe ser >= 0", "precioVehiculo"))
	}
	if in.PorcentajeCuotaInicial < 0 || in.PorcentajeCuotaInicial > 100 {
		errs = append(errs, domain.NewError("validation_error", "porcentajeCuotaInicial debe estar entre 0 y 100", "porcentajeCuotaInicial"))
	}
	if in.PlazoMeses < 1 {
		errs = append(errs, domain.NewError("validation_error", "plazoMeses debe ser >= 1", "plazoMeses"))
	}
	if in.PeriodosPorAnio < 1 {
		errs = append(errs, domain.NewError("validation_error", "periodosPorAnio debe ser >= 1", "periodosPorAnio"))
	}
	if in.PeriodosGracia < 0 {
		errs = append(errs, domain.NewError("validation_error", "periodosGracia debe ser >= 0", "periodosGracia"))
	}
	if in.TipoGracia != domain.GraceTotal && in.TipoGracia != domain.GraceParcial {
		errs = append(errs, domain.NewError("validation_error", "tipoGracia debe ser total o parcial", "tipoGracia"))
	}
	if in.ValorFinal < 0 {
		errs = append(errs, domain.NewError("validation_error", "valorFinal debe ser >= 0", "valorFinal"))
	}
	if in.CostosFinanciados < 0 {
		errs = append(errs, domain.NewError("validation_error", "costosFinanciados debe ser >= 0", "costosFinanciados"))
	}
	if in.CostosIniciales < 0 {
		errs = append(errs, domain.NewError("validation_error", "costosIniciales debe ser >= 0", "costosIniciales"))
	}
	return errs
}

func CalculateSimulation(in domain.SimulacionInput) domain.SimulacionResult {
	tea := in.TasaEfectivaAnual
	if tea > 1 {
		tea = tea / 100
	}
	periodosPorAnio := in.PeriodosPorAnio
	if periodosPorAnio <= 0 {
		periodosPorAnio = 12
	}
	i := math.Pow(1+tea, 1/float64(periodosPorAnio)) - 1
	montoFinanciado := in.PrecioVehiculo - in.PrecioVehiculo*(in.PorcentajeCuotaInicial/100) + in.CostosFinanciados
	montoNeto := montoFinanciado - in.CostosIniciales
	gracia := in.PeriodosGracia
	if gracia > in.PlazoMeses {
		gracia = in.PlazoMeses
	}
	remaining := in.PlazoMeses - gracia
	saldo := montoFinanciado
	cron := make([]domain.Pago, 0, in.PlazoMeses)

	for mes := 1; mes <= gracia; mes++ {
		interes := saldo * i
		if in.TipoGracia == domain.GraceTotal {
			saldo += interes
			cron = append(cron, domain.Pago{Mes: mes, Cuota: 0, Interes: util.Round2(interes), Amortizacion: 0, SaldoDeudor: util.Round2(saldo)})
		} else {
			cron = append(cron, domain.Pago{Mes: mes, Cuota: util.Round2(interes), Interes: util.Round2(interes), Amortizacion: 0, SaldoDeudor: util.Round2(saldo)})
		}
	}

	cuotaBase := 0.0
	if remaining > 0 {
		if i == 0 {
			amortNoBalloon := (saldo - in.ValorFinal)
			if amortNoBalloon < 0 {
				amortNoBalloon = 0
			}
			cuotaBase = amortNoBalloon / float64(remaining)
		} else {
			factor := (i * math.Pow(1+i, float64(remaining))) / (math.Pow(1+i, float64(remaining)) - 1)
			cuotaBase = (saldo - (in.ValorFinal / math.Pow(1+i, float64(remaining)))) * factor
		}
	}

	for k := 1; k <= remaining; k++ {
		mes := gracia + k
		interes := saldo * i
		cuota := cuotaBase
		if k == remaining {
			cuota += in.ValorFinal
		}
		amort := cuota - interes
		if k == remaining || amort > saldo {
			amort = saldo
			cuota = interes + amort
		}
		saldo -= amort
		if saldo < 1e-8 {
			saldo = 0
		}
		cron = append(cron, domain.Pago{Mes: mes, Cuota: util.Round2(cuota), Interes: util.Round2(interes), Amortizacion: util.Round2(amort), SaldoDeudor: util.Round2(saldo)})
	}

	flows := make([]float64, 0, len(cron)+1)
	flows = append(flows, montoNeto)
	costoTotal := 0.0
	for _, p := range cron {
		flows = append(flows, -p.Cuota)
		costoTotal += p.Cuota
	}

	van := npv(i, flows)
	tir := irr(flows)
	tcea := -1.0
	if tir > -1 {
		tcea = math.Pow(1+tir, float64(periodosPorAnio)) - 1
	}

	return domain.SimulacionResult{
		TasaPeriodo:     i,
		CuotaBase:       util.Round2(cuotaBase),
		VAN:             van,
		TIR:             tir,
		TCEA:            tcea,
		CostoTotal:      util.Round2(costoTotal),
		MontoNeto:       util.Round2(montoNeto),
		MontoFinanciado: util.Round2(montoFinanciado),
		Cronograma:      cron,
	}
}

func npv(r float64, flows []float64) float64 {
	sum := 0.0
	for t, f := range flows {
		sum += f / math.Pow(1+r, float64(t))
	}
	return sum
}

func irr(flows []float64) float64 {
	x := 0.1
	for n := 0; n < 100; n++ {
		fx := 0.0
		dfx := 0.0
		for t, f := range flows {
			den := math.Pow(1+x, float64(t))
			fx += f / den
			if t > 0 {
				dfx -= float64(t) * f / math.Pow(1+x, float64(t+1))
			}
		}
		if math.Abs(fx) < 1e-10 {
			return x
		}
		if dfx == 0 {
			break
		}
		nx := x - fx/dfx
		if nx <= -0.9999 || math.IsNaN(nx) || math.IsInf(nx, 0) {
			break
		}
		x = nx
	}
	lo, hi := -0.99, 10.0
	fLo := npv(lo, flows)
	fHi := npv(hi, flows)
	if fLo*fHi > 0 {
		return x
	}
	for i := 0; i < 200; i++ {
		mid := (lo + hi) / 2
		fMid := npv(mid, flows)
		if math.Abs(fMid) < 1e-10 {
			return mid
		}
		if fLo*fMid < 0 {
			hi = mid
			fHi = fMid
		} else {
			lo = mid
			fLo = fMid
		}
		_ = fHi
	}
	return (lo + hi) / 2
}
