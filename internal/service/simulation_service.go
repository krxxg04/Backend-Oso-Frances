package service

import (
	"backend-of/internal/domain"
	"backend-of/internal/repository"
	"backend-of/internal/util"
	"context"
	"errors"
	"math"
	"strings"
	"time"
)

type SimulationService struct {
	clientes repository.ClienteRepository
	sims     repository.SimulacionRepository
}

func NewSimulationService(clientes repository.ClienteRepository, sims repository.SimulacionRepository) *SimulationService {
	return &SimulationService{clientes: clientes, sims: sims}
}

func ListMockBanks() []domain.Banco {
	options := BankOptions()
	out := make([]domain.Banco, 0, len(options))
	for _, option := range options {
		out = append(out, domain.Banco{
			ID:             option.ID,
			Nombre:         option.Nombre,
			TEAReferencial: option.TasaEfectivaAnual,
		})
	}
	return out
}

func ApplySimulationRules(in domain.SimulacionInput) (domain.SimulacionInput, []domain.APIError) {
	in = normalizeSimulationInput(in)
	return in, ValidateSimulationInput(in)
}

func (s *SimulationService) CreateForUser(ctx context.Context, username string, in domain.SimulacionInput) (domain.Simulacion, error) {
	in.NombreCliente = username
	in, errs := ApplySimulationRules(in)
	if len(errs) > 0 {
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
	return s.ListByUserFiltered(ctx, username, domain.SimulacionFilter{})
}

func (s *SimulationService) ListByUserFiltered(ctx context.Context, username string, filter domain.SimulacionFilter) ([]domain.Simulacion, error) {
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
		for _, item := range items {
			if matchesSimulationFilter(item, filter) {
				out = append(out, item)
			}
		}
	}
	return out, nil
}

func (s *SimulationService) GetByID(ctx context.Context, id string) (domain.Simulacion, bool, error) {
	return s.sims.GetByID(ctx, id)
}

func (s *SimulationService) GetByIDForUser(ctx context.Context, username, id string) (domain.Simulacion, bool, error) {
	sim, ok, err := s.sims.GetByID(ctx, id)
	if err != nil || !ok {
		return sim, ok, err
	}
	clientes, err := s.clientes.GetByNombre(ctx, username)
	if err != nil {
		return domain.Simulacion{}, false, err
	}
	for _, c := range clientes {
		if c.ID == sim.ClienteID {
			return sim, true, nil
		}
	}
	return domain.Simulacion{}, false, nil
}

func ValidateSimulationInput(in domain.SimulacionInput) []domain.APIError {
	in = normalizeSimulationInput(in)
	errs := make([]domain.APIError, 0)
	if util.NormalizeKey(in.NombreCliente) == "" {
		errs = append(errs, domain.NewError("validation_error", "nombreCliente es requerido", "nombreCliente"))
	}
	if in.Moneda != "" && in.Moneda != domain.CurrencyPEN && in.Moneda != domain.CurrencyUSD {
		errs = append(errs, domain.NewError("validation_error", "moneda debe ser PEN o USD", "moneda"))
	}
	if in.BancoID != "" {
		option, ok := FindBankOption(in.BancoID)
		if !ok {
			errs = append(errs, domain.NewError("validation_error", "bancoId no existe en el catalogo de bancos", "bancoId"))
		} else {
			if !containsTerm(option.PlazosMeses, in.PlazoMeses) {
				errs = append(errs, domain.NewError("validation_error", "plazoMeses no esta disponible para el banco seleccionado", "plazoMeses"))
			}
			if option.MontoMin > 0 && in.PrecioVehiculo < option.MontoMin {
				errs = append(errs, domain.NewError("validation_error", "precioVehiculo es menor al monto minimo del banco seleccionado", "precioVehiculo"))
			}
			if option.MontoMax > 0 && in.PrecioVehiculo > option.MontoMax {
				errs = append(errs, domain.NewError("validation_error", "precioVehiculo supera el monto maximo del banco seleccionado", "precioVehiculo"))
			}
			if in.PorcentajeCuotaInicial < option.PorcentajeCuotaInicialMin || (option.PorcentajeCuotaInicialMax > 0 && in.PorcentajeCuotaInicial > option.PorcentajeCuotaInicialMax) {
				errs = append(errs, domain.NewError("validation_error", "porcentajeCuotaInicial esta fuera del rango del banco seleccionado", "porcentajeCuotaInicial"))
			}
			if option.PeriodosGraciaMax > 0 && in.PeriodosGracia > option.PeriodosGraciaMax {
				errs = append(errs, domain.NewError("validation_error", "periodosGracia supera el maximo del banco seleccionado", "periodosGracia"))
			}
		}
	}
	if in.PrecioVehiculo < 0 {
		errs = append(errs, domain.NewError("validation_error", "precioVehiculo debe ser >= 0", "precioVehiculo"))
	}
	if in.Vehiculo.Precio < 0 {
		errs = append(errs, domain.NewError("validation_error", "vehiculo.precio debe ser >= 0", "vehiculo.precio"))
	}
	if in.PorcentajeCuotaInicial < 0 || in.PorcentajeCuotaInicial > 100 {
		errs = append(errs, domain.NewError("validation_error", "porcentajeCuotaInicial debe estar entre 0 y 100", "porcentajeCuotaInicial"))
	}
	if in.PlazoMeses < 1 {
		errs = append(errs, domain.NewError("validation_error", "plazoMeses debe ser >= 1", "plazoMeses"))
	}
	if in.PlazoMeses != 24 && in.PlazoMeses != 36 {
		errs = append(errs, domain.NewError("validation_error", "plazoMeses debe ser 24 o 36 para Compra Inteligente", "plazoMeses"))
	}
	if in.TasaAnual < 0 || in.TasaEfectivaAnual < 0 {
		errs = append(errs, domain.NewError("validation_error", "la tasa efectiva anual debe ser >= 0", "tasaEfectivaAnual"))
	}
	if in.PeriodosPorAnio < 1 {
		errs = append(errs, domain.NewError("validation_error", "periodosPorAnio debe ser >= 1", "periodosPorAnio"))
	}
	if in.PeriodosGracia < 0 {
		errs = append(errs, domain.NewError("validation_error", "periodosGracia debe ser >= 0", "periodosGracia"))
	}
	if in.TipoGracia != "" && in.TipoGracia != domain.GraceNone && in.TipoGracia != domain.GraceTotal && in.TipoGracia != domain.GraceParcial {
		errs = append(errs, domain.NewError("validation_error", "tipoGracia debe ser sin_gracia, total o parcial", "tipoGracia"))
	}
	if in.TipoGracia == domain.GraceNone && in.PeriodosGracia > 0 {
		errs = append(errs, domain.NewError("validation_error", "periodosGracia debe ser 0 cuando tipoGracia es sin_gracia", "periodosGracia"))
	}
	if (in.TipoGracia == domain.GraceTotal || in.TipoGracia == domain.GraceParcial) && in.PeriodosGracia == 0 {
		errs = append(errs, domain.NewError("validation_error", "periodosGracia debe ser mayor a 0 para gracia total o parcial", "periodosGracia"))
	}
	if in.ValorFinal < 0 {
		errs = append(errs, domain.NewError("validation_error", "valorFinal debe ser >= 0", "valorFinal"))
	}
	if in.CuotaFinalBalloon < 0 {
		errs = append(errs, domain.NewError("validation_error", "cuotaFinalBalloon debe ser >= 0", "cuotaFinalBalloon"))
	}
	if in.SeguroVehicularMensual < 0 || in.SeguroDesgravamenAnual < 0 {
		errs = append(errs, domain.NewError("validation_error", "los seguros deben ser >= 0", "seguros"))
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
	in = normalizeSimulationInput(in)
	bankOption := bankOptionForSimulation(in)
	tea := annualEffectiveRate(in)
	periodosPorAnio := in.PeriodosPorAnio
	if periodosPorAnio <= 0 {
		periodosPorAnio = 12
	}
	i := math.Pow(1+tea, 1/float64(periodosPorAnio)) - 1
	montoFinanciado := in.PrecioVehiculo - in.PrecioVehiculo*(in.PorcentajeCuotaInicial/100) + in.CostosFinanciados
	cuotaInicial := in.PrecioVehiculo * (in.PorcentajeCuotaInicial / 100)
	montoNeto := montoFinanciado - in.CostosIniciales
	gracia := in.PeriodosGracia
	if gracia > in.PlazoMeses {
		gracia = in.PlazoMeses
	}
	remaining := in.PlazoMeses - gracia
	saldo := montoFinanciado
	cron := make([]domain.Pago, 0, in.PlazoMeses)
	fechaInicio, hasFecha := parseStartDate(in.FechaInicio)
	seguroDesgravamenPeriodo := periodicRate(in.SeguroDesgravamenAnual, periodosPorAnio)
	totalIntereses := 0.0
	totalSeguros := 0.0

	for mes := 1; mes <= gracia; mes++ {
		saldoInicial := saldo
		interes := saldo * i
		seguroVehicular := in.SeguroVehicularMensual
		seguroDesgravamen := saldoInicial * seguroDesgravamenPeriodo
		seguro := seguroVehicular + seguroDesgravamen
		totalIntereses += interes
		totalSeguros += seguro
		if in.TipoGracia == domain.GraceTotal {
			saldo += interes
			cron = append(cron, buildPago(mes, fechaInicio, hasFecha, saldoInicial, seguro, seguroVehicular, seguroDesgravamen, 0, interes, 0, saldo, in.TipoGracia))
		} else {
			cron = append(cron, buildPago(mes, fechaInicio, hasFecha, saldoInicial, seguro, seguroVehicular, seguroDesgravamen, interes, interes, 0, saldo, in.TipoGracia))
		}
	}

	cuotaBase := 0.0
	valorFinal := in.ValorFinal
	if remaining > 0 {
		if i == 0 {
			amortNoBalloon := saldo - valorFinal
			if amortNoBalloon < 0 {
				amortNoBalloon = 0
			}
			cuotaBase = amortNoBalloon / float64(remaining)
		} else {
			factor := (i * math.Pow(1+i, float64(remaining))) / (math.Pow(1+i, float64(remaining)) - 1)
			cuotaBase = (saldo - (valorFinal / math.Pow(1+i, float64(remaining)))) * factor
		}
	}

	for k := 1; k <= remaining; k++ {
		mes := gracia + k
		saldoInicial := saldo
		interes := saldo * i
		cuota := cuotaBase
		if k == remaining {
			cuota += valorFinal
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
		seguroVehicular := in.SeguroVehicularMensual
		seguroDesgravamen := saldoInicial * seguroDesgravamenPeriodo
		seguro := seguroVehicular + seguroDesgravamen
		totalIntereses += interes
		totalSeguros += seguro
		cron = append(cron, buildPago(mes, fechaInicio, hasFecha, saldoInicial, seguro, seguroVehicular, seguroDesgravamen, cuota, interes, amort, saldo, domain.GraceNone))
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
	fechaFinalizacion := ""
	if hasFecha && len(cron) > 0 {
		fechaFinalizacion = cron[len(cron)-1].Fecha
	}
	resumen := domain.FinancialSummary{
		MontoFinanciado:   util.Round2(montoFinanciado),
		CuotaInicial:      util.Round2(cuotaInicial),
		CuotaMensual:      util.Round2(cuotaBase + in.SeguroVehicularMensual + montoFinanciado*seguroDesgravamenPeriodo),
		CuotaFinalBalloon: util.Round2(valorFinal),
		TotalIntereses:    util.Round2(totalIntereses),
		TotalSeguros:      util.Round2(totalSeguros),
		TotalPagado:       util.Round2(costoTotal),
		VAN:               van,
		TIR:               tir,
		TCEA:              tcea,
		FechaFinalizacion: fechaFinalizacion,
	}

	return domain.SimulacionResult{
		Banco:           bankOption,
		TasaPeriodo:     i,
		Tasa:            domain.Rate{TasaEfectivaAnual: tea, PeriodosPagoPorAnio: periodosPorAnio},
		Seguros:         domain.Insurance{SeguroVehicularMensual: in.SeguroVehicularMensual, SeguroDesgravamenAnual: in.SeguroDesgravamenAnual},
		CuotaBase:       util.Round2(cuotaBase),
		VAN:             van,
		TIR:             tir,
		TCEA:            tcea,
		CostoTotal:      util.Round2(costoTotal),
		TotalIntereses:  util.Round2(totalIntereses),
		TotalSeguros:    util.Round2(totalSeguros),
		CuotaInicial:    util.Round2(cuotaInicial),
		MontoNeto:       util.Round2(montoNeto),
		MontoFinanciado: util.Round2(montoFinanciado),
		Resumen:         resumen,
		Cronograma:      cron,
	}
}

func normalizeSimulationInput(in domain.SimulacionInput) domain.SimulacionInput {
	if in.Moneda == "" {
		in.Moneda = domain.CurrencyPEN
	}
	if in.PrecioVehiculo == 0 && in.Vehiculo.Precio > 0 {
		in.PrecioVehiculo = in.Vehiculo.Precio
	}
	if in.Vehiculo.Precio == 0 {
		in.Vehiculo.Precio = in.PrecioVehiculo
	}
	if in.Vehiculo.Moneda == "" {
		in.Vehiculo.Moneda = in.Moneda
	}
	in = applyBankOption(in)
	if in.PeriodosPorAnio <= 0 {
		in.PeriodosPorAnio = 12
	}
	if in.TasaAnual == 0 {
		in.TasaAnual = in.TasaEfectivaAnual
	}
	if in.TipoGracia == "" {
		in.TipoGracia = domain.GraceNone
	}
	if in.CuotaFinalBalloon > 0 {
		in.ValorFinal = in.CuotaFinalBalloon
	} else {
		in.CuotaFinalBalloon = in.ValorFinal
	}
	return in
}

func annualEffectiveRate(in domain.SimulacionInput) float64 {
	rate := in.TasaAnual
	if rate == 0 {
		rate = in.TasaEfectivaAnual
	}
	if rate > 1 {
		rate = rate / 100
	}
	return rate
}

func periodicRate(annual float64, periodsPerYear int) float64 {
	if annual <= 0 {
		return 0
	}
	if annual > 1 {
		annual = annual / 100
	}
	if periodsPerYear <= 0 {
		periodsPerYear = 12
	}
	return math.Pow(1+annual, 1/float64(periodsPerYear)) - 1
}

func annualEffectiveFromMonthlyPercent(monthlyPercent float64) float64 {
	if monthlyPercent <= 0 {
		return 0
	}
	monthlyRate := monthlyPercent / 100.0
	return math.Pow(1+monthlyRate, 12) - 1
}

func parseStartDate(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", raw)
	return t, err == nil
}

func buildPago(mes int, fechaInicio time.Time, hasFecha bool, saldoInicial, seguro, seguroVehicular, seguroDesgravamen, cuotaCapitalInteres, interes, amortizacion, saldoFinal float64, tipoGracia domain.GraceType) domain.Pago {
	fechaPago := time.Time{}
	fecha := ""
	if hasFecha {
		fechaPago = fechaInicio.AddDate(0, mes, 0)
		fecha = fechaPago.Format("2006-01-02")
	}
	cuotaTotal := cuotaCapitalInteres + seguro
	return domain.Pago{
		Mes:                 mes,
		Periodo:             mes,
		Fecha:               fecha,
		FechaPago:           fechaPago,
		SaldoInicial:        util.Round2(saldoInicial),
		Cuota:               util.Round2(cuotaTotal),
		CuotaCapitalInteres: util.Round2(cuotaCapitalInteres),
		Interes:             util.Round2(interes),
		SeguroVehicular:     util.Round2(seguroVehicular),
		SeguroDesgravamen:   util.Round2(seguroDesgravamen),
		Seguro:              util.Round2(seguro),
		Amortizacion:        util.Round2(amortizacion),
		SaldoFinal:          util.Round2(saldoFinal),
		SaldoDeudor:         util.Round2(saldoFinal),
		TipoGracia:          tipoGracia,
	}
}

func matchesSimulationFilter(sim domain.Simulacion, filter domain.SimulacionFilter) bool {
	if filter.Moneda != "" && sim.Input.Moneda != filter.Moneda {
		return false
	}
	if filter.PlazoMeses > 0 && sim.Input.PlazoMeses != filter.PlazoMeses {
		return false
	}
	if filter.MontoMin > 0 && sim.Result.MontoFinanciado < filter.MontoMin {
		return false
	}
	if filter.MontoMax > 0 && sim.Result.MontoFinanciado > filter.MontoMax {
		return false
	}
	if filter.FechaDesde != "" {
		from, err := time.Parse("2006-01-02", filter.FechaDesde)
		if err == nil && sim.CreadoEn.Before(from) {
			return false
		}
	}
	if filter.FechaHasta != "" {
		to, err := time.Parse("2006-01-02", filter.FechaHasta)
		if err == nil && sim.CreadoEn.After(to.AddDate(0, 0, 1)) {
			return false
		}
	}
	if filter.Vehiculo != "" {
		needle := util.NormalizeKey(filter.Vehiculo)
		haystack := util.NormalizeKey(strings.TrimSpace(sim.Input.Vehiculo.Marca + " " + sim.Input.Vehiculo.Modelo + " " + sim.Input.Vehiculo.Tipo))
		if !strings.Contains(haystack, needle) {
			return false
		}
	}
	return true
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
