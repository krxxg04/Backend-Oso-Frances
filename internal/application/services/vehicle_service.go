package services

import (
	domain "backend-of/internal/domain/entities"
	"backend-of/internal/domain/repositories"
	"backend-of/internal/domain/shared"
	"context"
	"errors"
)

type VehicleService struct {
	users    repository.UserRepository
	vehicles repository.VehicleRepository
}

func NewVehicleService(users repository.UserRepository, vehicles repository.VehicleRepository) *VehicleService {
	return &VehicleService{users: users, vehicles: vehicles}
}

func (s *VehicleService) CreateForUser(ctx context.Context, username string, vehicle domain.Vehicle) (domain.Vehicle, error) {
	if errs := ValidateVehicle(vehicle); len(errs) > 0 {
		return domain.Vehicle{}, errors.New("validation_error")
	}
	vehicle = normalizeVehicle(vehicle)
	user, ok, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return domain.Vehicle{}, err
	}
	if !ok {
		return domain.Vehicle{}, errors.New("unauthorized")
	}
	vehicle.UserID = user.ID
	return s.vehicles.CreateVehicle(ctx, vehicle)
}

func (s *VehicleService) ListByUser(ctx context.Context, username string) ([]domain.Vehicle, error) {
	user, ok, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("unauthorized")
	}
	return s.vehicles.ListVehiclesByUserID(ctx, user.ID)
}

func (s *VehicleService) GetByIDForUser(ctx context.Context, username, id string) (domain.Vehicle, bool, error) {
	vehicle, ok, err := s.vehicles.GetVehicleByID(ctx, id)
	if err != nil || !ok {
		return vehicle, ok, err
	}
	user, ok, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return domain.Vehicle{}, false, err
	}
	if !ok {
		return domain.Vehicle{}, false, nil
	}
	return vehicle, vehicle.UserID == user.ID, nil
}

func (s *VehicleService) UpdateForUser(ctx context.Context, username, id string, vehicle domain.Vehicle) (domain.Vehicle, bool, error) {
	if errs := ValidateVehicle(vehicle); len(errs) > 0 {
		return domain.Vehicle{}, false, errors.New("validation_error")
	}
	user, ok, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return domain.Vehicle{}, false, err
	}
	if !ok {
		return domain.Vehicle{}, false, nil
	}
	existing, ok, err := s.vehicles.GetVehicleByID(ctx, id)
	if err != nil {
		return domain.Vehicle{}, false, err
	}
	if !ok || existing.UserID != user.ID {
		return domain.Vehicle{}, false, nil
	}
	vehicle = normalizeVehicle(vehicle)
	vehicle.ID = existing.ID
	vehicle.UserID = existing.UserID
	vehicle.CreadoEn = existing.CreadoEn
	updated, ok, err := s.vehicles.UpdateVehicle(ctx, vehicle)
	return updated, ok, err
}

func ValidateVehicle(vehicle domain.Vehicle) []domain.APIError {
	vehicle = normalizeVehicle(vehicle)
	errs := make([]domain.APIError, 0)
	if shared.NormalizeKey(vehicle.Marca) == "" {
		errs = append(errs, domain.NewError("validation_error", "marca es requerida", "marca"))
	}
	if shared.NormalizeKey(vehicle.Modelo) == "" {
		errs = append(errs, domain.NewError("validation_error", "modelo es requerido", "modelo"))
	}
	if vehicle.Anio != 0 && (vehicle.Anio < 1990 || vehicle.Anio > 2100) {
		errs = append(errs, domain.NewError("validation_error", "anio debe estar entre 1990 y 2100", "anio"))
	}
	if vehicle.Precio < minimumVehicleAmount {
		errs = append(errs, domain.NewError("validation_error", "precio debe ser >= 2000", "precio"))
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
