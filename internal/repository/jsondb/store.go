package jsondb

import (
	"backend-of/internal/domain"
	"backend-of/internal/util"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type dbData struct {
	Users         map[string]domain.User       `json:"users"`
	Clientes      map[string]domain.Cliente    `json:"clientes"`
	ClientesByKey map[string]string            `json:"clientesByKey"`
	Vehicles      map[string]domain.Vehicle    `json:"vehicles"`
	Simulaciones  map[string]domain.Simulacion `json:"simulaciones"`
	Seq           int64                        `json:"seq"`
}

type Store struct {
	mu   sync.RWMutex
	path string
	data dbData
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = dbData{
		Users:         make(map[string]domain.User),
		Clientes:      make(map[string]domain.Cliente),
		ClientesByKey: make(map[string]string),
		Vehicles:      make(map[string]domain.Vehicle),
		Simulaciones:  make(map[string]domain.Simulacion),
	}

	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s.persistLocked()
		}
		return err
	}
	if len(b) == 0 {
		return nil
	}
	if err := json.Unmarshal(b, &s.data); err != nil {
		return err
	}
	s.ensureMapsLocked()
	return nil
}

func (s *Store) ensureMapsLocked() {
	if s.data.Users == nil {
		s.data.Users = make(map[string]domain.User)
	}
	if s.data.Clientes == nil {
		s.data.Clientes = make(map[string]domain.Cliente)
	}
	if s.data.ClientesByKey == nil {
		s.data.ClientesByKey = make(map[string]string)
	}
	if s.data.Vehicles == nil {
		s.data.Vehicles = make(map[string]domain.Vehicle)
	}
	if s.data.Simulaciones == nil {
		s.data.Simulaciones = make(map[string]domain.Simulacion)
	}
}

func (s *Store) persistLocked() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}

func (s *Store) nextID(prefix string) string {
	s.data.Seq++
	return prefix + "_" + time.Now().UTC().Format("20060102150405") + "_" + strconv.FormatInt(s.data.Seq, 10)
}

func (s *Store) SeedIfEmpty(_ context.Context, users []domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.data.Users) > 0 {
		return nil
	}
	for _, u := range users {
		s.data.Users[u.Username] = u
	}
	return s.persistLocked()
}

func (s *Store) CreateUser(_ context.Context, user domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.Users[user.Username]; exists {
		return fmt.Errorf("user_exists")
	}
	s.data.Users[user.Username] = user
	return s.persistLocked()
}

func (s *Store) GetByUsername(_ context.Context, username string) (domain.User, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.data.Users[username]
	return u, ok, nil
}

func (s *Store) GetOrCreateByNombre(_ context.Context, nombre string) (domain.Cliente, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := util.NormalizeKey(nombre)
	if id, ok := s.data.ClientesByKey[key]; ok {
		return s.data.Clientes[id], nil
	}
	id := s.nextID("cli")
	c := domain.Cliente{ID: id, Nombre: nombre, NombreKey: key, CreadoEn: time.Now().UTC()}
	s.data.Clientes[id] = c
	s.data.ClientesByKey[key] = id
	return c, s.persistLocked()
}

func (s *Store) GetByNombre(_ context.Context, nombre string) ([]domain.Cliente, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := util.NormalizeKey(nombre)
	out := make([]domain.Cliente, 0)
	if key == "" {
		for _, c := range s.data.Clientes {
			out = append(out, c)
		}
	} else {
		if id, ok := s.data.ClientesByKey[key]; ok {
			out = append(out, s.data.Clientes[id])
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreadoEn.Before(out[j].CreadoEn) })
	return out, nil
}

func (s *Store) Create(_ context.Context, sim domain.Simulacion) (domain.Simulacion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sim.ID = s.nextID("sim")
	sim.CreadoEn = time.Now().UTC()
	s.data.Simulaciones[sim.ID] = sim
	return sim, s.persistLocked()
}

func (s *Store) GetByID(_ context.Context, id string) (domain.Simulacion, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sim, ok := s.data.Simulaciones[id]
	return sim, ok, nil
}

func (s *Store) ListByClienteID(_ context.Context, clienteID string) ([]domain.Simulacion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Simulacion, 0)
	for _, item := range s.data.Simulaciones {
		if item.ClienteID == clienteID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreadoEn.After(out[j].CreadoEn) })
	return out, nil
}

func (s *Store) CreateVehicle(_ context.Context, vehicle domain.Vehicle) (domain.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vehicle.ID = s.nextID("veh")
	vehicle.CreadoEn = time.Now().UTC()
	s.data.Vehicles[vehicle.ID] = vehicle
	return vehicle, s.persistLocked()
}

func (s *Store) GetVehicleByID(_ context.Context, id string) (domain.Vehicle, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vehicle, ok := s.data.Vehicles[id]
	return vehicle, ok, nil
}

func (s *Store) ListVehiclesByClienteID(_ context.Context, clienteID string) ([]domain.Vehicle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Vehicle, 0)
	for _, item := range s.data.Vehicles {
		if item.ClienteID == clienteID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreadoEn.After(out[j].CreadoEn) })
	return out, nil
}
