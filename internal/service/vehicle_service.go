package service

import (
	"backend-of/internal/domain"
	"backend-of/internal/repository"
	"backend-of/internal/util"
	"context"
	"errors"
)

type VehicleService struct {
	clientes repository.ClienteRepository
	vehicles repository.VehicleRepository
}

func NewVehicleService(clientes repository.ClienteRepository, vehicles repository.VehicleRepository) *VehicleService {
	return &VehicleService{clientes: clientes, vehicles: vehicles}
}

func (s *VehicleService) CreateForUser(ctx context.Context, username string, vehicle domain.Vehicle) (domain.Vehicle, error) {
	if errs := ValidateVehicle(vehicle); len(errs) > 0 {
		return domain.Vehicle{}, errors.New("validation_error")
	}
	vehicle = normalizeVehicle(vehicle)
	c, err := s.clientes.GetOrCreateByNombre(ctx, username)
	if err != nil {
		return domain.Vehicle{}, err
	}
	vehicle.ClienteID = c.ID
	return s.vehicles.CreateVehicle(ctx, vehicle)
}

func (s *VehicleService) ListByUser(ctx context.Context, username string) ([]domain.Vehicle, error) {
	clientes, err := s.clientes.GetByNombre(ctx, username)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Vehicle, 0)
	for _, c := range clientes {
		items, err := s.vehicles.ListVehiclesByClienteID(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

func (s *VehicleService) GetByIDForUser(ctx context.Context, username, id string) (domain.Vehicle, bool, error) {
	vehicle, ok, err := s.vehicles.GetVehicleByID(ctx, id)
	if err != nil || !ok {
		return vehicle, ok, err
	}
	clientes, err := s.clientes.GetByNombre(ctx, username)
	if err != nil {
		return domain.Vehicle{}, false, err
	}
	for _, c := range clientes {
		if c.ID == vehicle.ClienteID {
			return vehicle, true, nil
		}
	}
	return domain.Vehicle{}, false, nil
}

func ValidateVehicle(vehicle domain.Vehicle) []domain.APIError {
	vehicle = normalizeVehicle(vehicle)
	errs := make([]domain.APIError, 0)
	if util.NormalizeKey(vehicle.Marca) == "" {
		errs = append(errs, domain.NewError("validation_error", "marca es requerida", "marca"))
	}
	if util.NormalizeKey(vehicle.Modelo) == "" {
		errs = append(errs, domain.NewError("validation_error", "modelo es requerido", "modelo"))
	}
	if vehicle.Anio != 0 && (vehicle.Anio < 1990 || vehicle.Anio > 2100) {
		errs = append(errs, domain.NewError("validation_error", "anio debe estar entre 1990 y 2100", "anio"))
	}
	if vehicle.Precio <= 0 {
		errs = append(errs, domain.NewError("validation_error", "precio debe ser mayor a 0", "precio"))
	}
	if vehicle.Moneda != domain.CurrencyPEN && vehicle.Moneda != domain.CurrencyUSD {
		errs = append(errs, domain.NewError("validation_error", "moneda debe ser PEN o USD", "moneda"))
	}
	return errs
}

func normalizeVehicle(vehicle domain.Vehicle) domain.Vehicle {
	if vehicle.Moneda == "" {
		vehicle.Moneda = domain.CurrencyPEN
	}
	return vehicle
}
