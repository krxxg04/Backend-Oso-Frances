package repository

import (
	"context"

	domain "backend-of/internal/domain/entities"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (domain.User, bool, error)
	CreateUser(ctx context.Context, user domain.User) error
	SeedIfEmpty(ctx context.Context, users []domain.User) error
	UpdateProfileByUsername(ctx context.Context, username string, update domain.UserProfileUpdate) (domain.User, bool, error)
}

type SimulacionRepository interface {
	Create(ctx context.Context, sim domain.Simulacion) (domain.Simulacion, error)
	GetByID(ctx context.Context, id string) (domain.Simulacion, bool, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Simulacion, error)
}

type VehicleRepository interface {
	CreateVehicle(ctx context.Context, vehicle domain.Vehicle) (domain.Vehicle, error)
	GetVehicleByID(ctx context.Context, id string) (domain.Vehicle, bool, error)
	ListVehiclesByUserID(ctx context.Context, userID string) ([]domain.Vehicle, error)
	UpdateVehicle(ctx context.Context, vehicle domain.Vehicle) (domain.Vehicle, bool, error)
}
