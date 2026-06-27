package jsondb

import (
	domain "backend-of/internal/domain/entities"
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
	Users        map[string]domain.User       `json:"users"`
	Vehicles     map[string]domain.Vehicle    `json:"vehicles"`
	Simulaciones map[string]domain.Simulacion `json:"simulaciones"`
	Seq          int64                        `json:"seq"`
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
		Users:        make(map[string]domain.User),
		Vehicles:     make(map[string]domain.Vehicle),
		Simulaciones: make(map[string]domain.Simulacion),
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
	if s.data.Vehicles == nil {
		s.data.Vehicles = make(map[string]domain.Vehicle)
	}
	if s.data.Simulaciones == nil {
		s.data.Simulaciones = make(map[string]domain.Simulacion)
	}
	for username, user := range s.data.Users {
		if user.ID == "" {
			user.ID = s.nextID("usr")
			if user.Username == "" {
				user.Username = username
			}
			s.data.Users[username] = user
		}
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
		if u.ID == "" {
			u.ID = s.nextID("usr")
		}
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
	if user.Email != "" {
		for _, existing := range s.data.Users {
			if existing.Email == user.Email {
				return fmt.Errorf("email_exists")
			}
		}
	}
	if user.ID == "" {
		user.ID = s.nextID("usr")
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

func (s *Store) GetByEmail(_ context.Context, email string) (domain.User, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.data.Users {
		if user.Email == email {
			return user, true, nil
		}
	}
	return domain.User{}, false, nil
}

func (s *Store) GetByGoogleID(_ context.Context, googleID string) (domain.User, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.data.Users {
		if user.GoogleID == googleID {
			return user, true, nil
		}
	}
	return domain.User{}, false, nil
}

func (s *Store) UpdateProfileByUsername(_ context.Context, username string, update domain.UserProfileUpdate) (domain.User, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.data.Users[username]
	if !ok {
		return domain.User{}, false, nil
	}
	if update.Email != "" {
		user.Email = update.Email
	}
	if update.DNI != "" {
		user.DNI = update.DNI
	}
	if update.FullName != "" {
		user.FullName = update.FullName
	}
	if update.PictureURL != "" {
		user.PictureURL = update.PictureURL
	}
	s.data.Users[username] = user
	return user, true, s.persistLocked()
}

func (s *Store) LinkGoogleAccount(_ context.Context, username, googleID, email, fullName, pictureURL string) (domain.User, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.data.Users[username]
	if !ok {
		return domain.User{}, false, nil
	}
	user.GoogleID = googleID
	if email != "" {
		user.Email = email
	}
	if fullName != "" {
		user.FullName = fullName
	}
	if pictureURL != "" {
		user.PictureURL = pictureURL
	}
	if user.Role == "" {
		user.Role = "user"
	}
	s.data.Users[username] = user
	return user, true, s.persistLocked()
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

func (s *Store) ListByUserID(_ context.Context, userID string) ([]domain.Simulacion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Simulacion, 0)
	for _, item := range s.data.Simulaciones {
		if item.UserID == userID {
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

func (s *Store) ListVehiclesByUserID(_ context.Context, userID string) ([]domain.Vehicle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Vehicle, 0)
	for _, item := range s.data.Vehicles {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreadoEn.After(out[j].CreadoEn) })
	return out, nil
}

func (s *Store) UpdateVehicle(_ context.Context, vehicle domain.Vehicle) (domain.Vehicle, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Vehicles[vehicle.ID]; !ok {
		return domain.Vehicle{}, false, nil
	}
	s.data.Vehicles[vehicle.ID] = vehicle
	return vehicle, true, s.persistLocked()
}
