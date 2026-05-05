package repository

import (
	"context"

	"backend-of/internal/domain"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (domain.User, bool, error)
	SeedIfEmpty(ctx context.Context, users []domain.User) error
}

type ClienteRepository interface {
	GetOrCreateByNombre(ctx context.Context, nombre string) (domain.Cliente, error)
	GetByNombre(ctx context.Context, nombre string) ([]domain.Cliente, error)
}

type SimulacionRepository interface {
	Create(ctx context.Context, sim domain.Simulacion) (domain.Simulacion, error)
	GetByID(ctx context.Context, id string) (domain.Simulacion, bool, error)
	ListByClienteID(ctx context.Context, clienteID string) ([]domain.Simulacion, error)
}
