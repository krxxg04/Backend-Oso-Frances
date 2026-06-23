package postgres

import (
	domain "backend-of/internal/domain/entities"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	s := &Store{pool: pool}
	if err := s.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	q := `
create table if not exists users (
  id uuid primary key default gen_random_uuid(),
  username text unique,
  email text unique,
  dni text,
  full_name text,
  password_hash text,
  google_id text unique,
  picture_url text,
  role text not null default 'user',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists vehicles (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade,
  marca text not null,
  modelo text not null,
  anio integer not null default 0,
  tipo text not null default '',
  precio numeric not null,
  moneda text not null,
  creado_en timestamptz not null
);

create table if not exists simulaciones (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade,
  vehicle_id uuid references vehicles(id) on delete set null,
  creado_en timestamptz not null,
  input_json jsonb not null,
  result_json jsonb not null
);

create index if not exists idx_simulaciones_user_creado
on simulaciones(user_id, creado_en desc);

create index if not exists idx_simulaciones_vehicle
on simulaciones(vehicle_id);

create index if not exists idx_vehicles_user_creado
on vehicles(user_id, creado_en desc);
`
	_, err := s.pool.Exec(ctx, q)
	return err
}

func (s *Store) SeedIfEmpty(ctx context.Context, users []domain.User) error {
	var count int
	if err := s.pool.QueryRow(ctx, `select count(*) from users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, u := range users {
		_, err := s.pool.Exec(ctx, `insert into users (username, email, dni, full_name, password_hash, role) values ($1,$2,$3,$4,$5,$6)`, u.Username, nullableText(u.Email), nullableText(u.DNI), nullableText(u.FullName), u.PasswordHash, u.Role)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateUser(ctx context.Context, user domain.User) error {
	_, exists, err := s.GetByUsername(ctx, user.Username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user_exists")
	}
	_, err = s.pool.Exec(ctx, `insert into users (username, email, dni, full_name, password_hash, role) values ($1,$2,$3,$4,$5,$6)`, user.Username, nullableText(user.Email), nullableText(user.DNI), nullableText(user.FullName), user.PasswordHash, user.Role)
	return err
}

func (s *Store) GetByUsername(ctx context.Context, username string) (domain.User, bool, error) {
	var u domain.User
	err := s.pool.QueryRow(ctx, `select id, username, coalesce(email,''), coalesce(dni,''), coalesce(full_name,''), coalesce(password_hash,''), coalesce(google_id,''), coalesce(picture_url,''), role from users where username=$1`, username).Scan(&u.ID, &u.Username, &u.Email, &u.DNI, &u.FullName, &u.PasswordHash, &u.GoogleID, &u.PictureURL, &u.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, err
	}
	return u, true, nil
}

func (s *Store) GetByEmail(ctx context.Context, email string) (domain.User, bool, error) {
	var u domain.User
	err := s.pool.QueryRow(ctx, `select id, username, coalesce(email,''), coalesce(dni,''), coalesce(full_name,''), coalesce(password_hash,''), coalesce(google_id,''), coalesce(picture_url,''), role from users where email=$1`, email).Scan(&u.ID, &u.Username, &u.Email, &u.DNI, &u.FullName, &u.PasswordHash, &u.GoogleID, &u.PictureURL, &u.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, err
	}
	return u, true, nil
}

func (s *Store) GetByGoogleID(ctx context.Context, googleID string) (domain.User, bool, error) {
	var u domain.User
	err := s.pool.QueryRow(ctx, `select id, username, coalesce(email,''), coalesce(dni,''), coalesce(full_name,''), coalesce(password_hash,''), coalesce(google_id,''), coalesce(picture_url,''), role from users where google_id=$1`, googleID).Scan(&u.ID, &u.Username, &u.Email, &u.DNI, &u.FullName, &u.PasswordHash, &u.GoogleID, &u.PictureURL, &u.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, err
	}
	return u, true, nil
}

func (s *Store) UpdateProfileByUsername(ctx context.Context, username string, update domain.UserProfileUpdate) (domain.User, bool, error) {
	query := `
update users
set
  email = coalesce($2, email),
  dni = coalesce($3, dni),
  full_name = coalesce($4, full_name),
  picture_url = coalesce($5, picture_url),
  updated_at = now()
where username = $1
returning id, username, coalesce(email,''), coalesce(dni,''), coalesce(full_name,''), coalesce(password_hash,''), coalesce(google_id,''), coalesce(picture_url,''), role`
	var user domain.User
	err := s.pool.QueryRow(ctx, query, username, nullableText(update.Email), nullableText(update.DNI), nullableText(update.FullName), nullableText(update.PictureURL)).
		Scan(&user.ID, &user.Username, &user.Email, &user.DNI, &user.FullName, &user.PasswordHash, &user.GoogleID, &user.PictureURL, &user.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, err
	}
	return user, true, nil
}

func (s *Store) LinkGoogleAccount(ctx context.Context, username, googleID, email, fullName, pictureURL string) (domain.User, bool, error) {
	query := `
update users
set
  google_id = coalesce($2, google_id),
  email = coalesce($3, email),
  full_name = coalesce($4, full_name),
  picture_url = coalesce($5, picture_url),
  role = 'user',
  updated_at = now()
where username = $1
returning id, username, coalesce(email,''), coalesce(dni,''), coalesce(full_name,''), coalesce(password_hash,''), coalesce(google_id,''), coalesce(picture_url,''), role`
	var user domain.User
	err := s.pool.QueryRow(ctx, query, username, nullableText(googleID), nullableText(email), nullableText(fullName), nullableText(pictureURL)).
		Scan(&user.ID, &user.Username, &user.Email, &user.DNI, &user.FullName, &user.PasswordHash, &user.GoogleID, &user.PictureURL, &user.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, err
	}
	return user, true, nil
}

func (s *Store) Create(ctx context.Context, sim domain.Simulacion) (domain.Simulacion, error) {
	sim.CreadoEn = time.Now().UTC()
	inJSON, err := json.Marshal(sim.Input)
	if err != nil {
		return domain.Simulacion{}, err
	}
	resJSON, err := json.Marshal(sim.Result)
	if err != nil {
		return domain.Simulacion{}, err
	}
	err = s.pool.QueryRow(ctx, `insert into simulaciones (user_id, vehicle_id, creado_en, input_json, result_json) values ($1,$2,$3,$4,$5) returning id`, sim.UserID, nullableUUID(sim.VehicleID), sim.CreadoEn, inJSON, resJSON).Scan(&sim.ID)
	if err != nil {
		return domain.Simulacion{}, err
	}
	return sim, nil
}

func (s *Store) GetByID(ctx context.Context, id string) (domain.Simulacion, bool, error) {
	var sim domain.Simulacion
	var inJSON, resJSON []byte
	err := s.pool.QueryRow(ctx, `select id,user_id,coalesce(vehicle_id::text,''),creado_en,input_json,result_json from simulaciones where id=$1`, id).Scan(&sim.ID, &sim.UserID, &sim.VehicleID, &sim.CreadoEn, &inJSON, &resJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Simulacion{}, false, nil
		}
		return domain.Simulacion{}, false, err
	}
	if err := json.Unmarshal(inJSON, &sim.Input); err != nil {
		return domain.Simulacion{}, false, err
	}
	if err := json.Unmarshal(resJSON, &sim.Result); err != nil {
		return domain.Simulacion{}, false, err
	}
	return sim, true, nil
}

func (s *Store) ListByUserID(ctx context.Context, userID string) ([]domain.Simulacion, error) {
	rows, err := s.pool.Query(ctx, `select id,user_id,coalesce(vehicle_id::text,''),creado_en,input_json,result_json from simulaciones where user_id=$1 order by creado_en desc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Simulacion, 0)
	for rows.Next() {
		var sim domain.Simulacion
		var inJSON, resJSON []byte
		if err := rows.Scan(&sim.ID, &sim.UserID, &sim.VehicleID, &sim.CreadoEn, &inJSON, &resJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(inJSON, &sim.Input); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(resJSON, &sim.Result); err != nil {
			return nil, err
		}
		out = append(out, sim)
	}
	return out, rows.Err()
}

func (s *Store) CreateVehicle(ctx context.Context, vehicle domain.Vehicle) (domain.Vehicle, error) {
	vehicle.CreadoEn = time.Now().UTC()
	err := s.pool.QueryRow(ctx, `
insert into vehicles (user_id,marca,modelo,anio,tipo,precio,moneda,creado_en)
values ($1,$2,$3,$4,$5,$6,$7,$8)
returning id`,
		vehicle.UserID, vehicle.Marca, vehicle.Modelo, vehicle.Anio, vehicle.Tipo, vehicle.Precio, vehicle.Moneda, vehicle.CreadoEn).Scan(&vehicle.ID)
	if err != nil {
		return domain.Vehicle{}, err
	}
	return vehicle, nil
}

func (s *Store) GetVehicleByID(ctx context.Context, id string) (domain.Vehicle, bool, error) {
	var vehicle domain.Vehicle
	err := s.pool.QueryRow(ctx, `
select id,user_id,marca,modelo,anio,tipo,precio,moneda,creado_en
from vehicles
where id=$1`, id).Scan(&vehicle.ID, &vehicle.UserID, &vehicle.Marca, &vehicle.Modelo, &vehicle.Anio, &vehicle.Tipo, &vehicle.Precio, &vehicle.Moneda, &vehicle.CreadoEn)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Vehicle{}, false, nil
		}
		return domain.Vehicle{}, false, err
	}
	return vehicle, true, nil
}

func (s *Store) ListVehiclesByUserID(ctx context.Context, userID string) ([]domain.Vehicle, error) {
	rows, err := s.pool.Query(ctx, `
select id,user_id,marca,modelo,anio,tipo,precio,moneda,creado_en
from vehicles
where user_id=$1
order by creado_en desc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Vehicle, 0)
	for rows.Next() {
		var vehicle domain.Vehicle
		if err := rows.Scan(&vehicle.ID, &vehicle.UserID, &vehicle.Marca, &vehicle.Modelo, &vehicle.Anio, &vehicle.Tipo, &vehicle.Precio, &vehicle.Moneda, &vehicle.CreadoEn); err != nil {
			return nil, err
		}
		out = append(out, vehicle)
	}
	return out, rows.Err()
}

func (s *Store) UpdateVehicle(ctx context.Context, vehicle domain.Vehicle) (domain.Vehicle, bool, error) {
	query := `
update vehicles
set marca=$2, modelo=$3, anio=$4, tipo=$5, precio=$6, moneda=$7
where id=$1 and user_id=$8
returning id,user_id,marca,modelo,anio,tipo,precio,moneda,creado_en`
	err := s.pool.QueryRow(ctx, query, vehicle.ID, vehicle.Marca, vehicle.Modelo, vehicle.Anio, vehicle.Tipo, vehicle.Precio, vehicle.Moneda, vehicle.UserID).
		Scan(&vehicle.ID, &vehicle.UserID, &vehicle.Marca, &vehicle.Modelo, &vehicle.Anio, &vehicle.Tipo, &vehicle.Precio, &vehicle.Moneda, &vehicle.CreadoEn)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Vehicle{}, false, nil
		}
		return domain.Vehicle{}, false, err
	}
	return vehicle, true, nil
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableUUID(value string) any {
	if value == "" {
		return nil
	}
	return value
}
