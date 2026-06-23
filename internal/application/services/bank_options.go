package services

import (
	domain "backend-of/internal/domain/entities"
	"math"
	"strings"
)

const minimumVehicleAmount = 2000.0

func BankOptions() []domain.BankOption {
	return []domain.BankOption{
		buildBankOption(domain.BankOption{
			ID:                        "bcp-compra-inteligente",
			Nombre:                    "BCP",
			Producto:                  "Credito Vehicular Compra Inteligente",
			Moneda:                    domain.CurrencyPEN,
			TasaEfectivaAnual:         12.35,
			SeguroDesgravamenMensual:  0.077,
			MontoMin:                  15000,
			PorcentajeCuotaInicialMin: 0,
			PorcentajeCuotaInicialMax: 50,
			PlazosMeses:               []int{24, 36},
			PeriodosGraciaMax:         3,
			Fuente:                    "BCP / Comparabien",
			FuenteURL:                 "https://www.viabcp.com/creditos/credito-vehicular/simulador-vehicular/",
			Notas:                     "Seguro desgravamen mensual referencial aplicado sobre saldo segun referencia comercial.",
		}),
		buildBankOption(domain.BankOption{
			ID:                        "bbva-vehicular-sostenible",
			Nombre:                    "BBVA",
			Producto:                  "Prestamo Vehicular Sostenible",
			Moneda:                    domain.CurrencyPEN,
			TasaEfectivaAnual:         11.49,
			SeguroDesgravamenMensual:  0.069,
			MontoMin:                  28800,
			PorcentajeCuotaInicialMin: 0,
			PorcentajeCuotaInicialMax: 50,
			PlazosMeses:               []int{24, 36},
			Fuente:                    "BBVA Peru",
			FuenteURL:                 "https://www.bbva.pe/personas/productos/prestamos/credito-vehicular/prestamo-vehicular-sostenible.html",
			Notas:                     "Seguro desgravamen mensual individual referencial de BBVA.",
		}),
		buildBankOption(domain.BankOption{
			ID:                        "scotiabank-vehicular",
			Nombre:                    "Scotiabank",
			Producto:                  "Credito Vehicular",
			Moneda:                    domain.CurrencyPEN,
			TasaEfectivaAnual:         15.99,
			SeguroDesgravamenMensual:  0.1045,
			SeguroVehicularMensualPct: 0.5064,
			MontoMin:                  28800,
			PorcentajeCuotaInicialMin: 0,
			PorcentajeCuotaInicialMax: 50,
			PlazosMeses:               []int{24, 36},
			PeriodosGraciaMax:         2,
			Fuente:                    "Scotiabank Peru",
			FuenteURL:                 "https://www.scotiabank.com.pe/Personas/Prestamos/Creditos/Vehicular",
			Notas:                     "Seguro desgravamen mensual referencial publicado para credito vehicular.",
		}),
	}
}

func FindBankOption(id string) (domain.BankOption, bool) {
	needle := normalizedBankID(id)
	if needle == "" {
		return domain.BankOption{}, false
	}
	for _, option := range BankOptions() {
		if strings.ToLower(option.ID) == needle {
			return option, true
		}
	}
	return domain.BankOption{}, false
}

func bankOptionForSimulation(in domain.SimulacionInput) *domain.BankOption {
	option, ok := FindBankOption(in.BancoID)
	if !ok {
		return nil
	}
	return &option
}

func applyBankOption(in domain.SimulacionInput) domain.SimulacionInput {
	option := bankOptionForSimulation(in)
	if option == nil {
		return in
	}
	in.Moneda = option.Moneda
	in.TasaAnual = option.TasaEfectivaAnual
	in.TasaEfectivaAnual = option.TasaEfectivaAnual
	in.SeguroDesgravamenAnual = option.SeguroDesgravamenAnual
	if in.SeguroVehicularMensual == 0 && option.SeguroVehicularMensualPct > 0 && in.PrecioVehiculo > 0 {
		in.SeguroVehicularMensual = in.PrecioVehiculo * (option.SeguroVehicularMensualPct / 100)
	}
	return in
}

func buildBankOption(option domain.BankOption) domain.BankOption {
	if option.SeguroDesgravamenAnual == 0 && option.SeguroDesgravamenMensual > 0 {
		option.SeguroDesgravamenAnual = monthlyPercentToAnnualPercent(option.SeguroDesgravamenMensual)
	}
	return option
}

func monthlyPercentToAnnualPercent(monthlyPercent float64) float64 {
	return math.Pow(1+monthlyPercent/100, 12) - 1
}

func containsTerm(terms []int, term int) bool {
	if len(terms) == 0 {
		return true
	}
	for _, item := range terms {
		if item == term {
			return true
		}
	}
	return false
}

func normalizedBankID(id string) string {
	needle := strings.TrimSpace(strings.ToLower(id))
	switch needle {
	case "", "manual", "sin-banco", "sin banco", "sin_banco", "none":
		return ""
	default:
		return needle
	}
}
