package postgres

import (
	"backend-of/internal/domain"
	"backend-of/internal/util"
	"context"
	"crypto/rand"
	"encoding/hex"
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
  username text primary key,
  email text,
  dni text,
  full_name text,
  password_hash text not null,
  role text not null,
  created_at timestamptz not null default now()
);

alter table users add column if not exists email text;
alter table users add column if not exists dni text;
alter table users add column if not exists full_name text;

create table if not exists clientes (
  id text primary key,
  nombre text not null,
  nombre_key text not null unique,
  creado_en timestamptz not null
);

create table if not exists simulaciones (
  id text primary key,
  cliente_id text not null references clientes(id),
  creado_en timestamptz not null,
  input_json jsonb not null,
  result_json jsonb not null
);

create index if not exists idx_sim_cliente_creado
on simulaciones(cliente_id, creado_en desc);
`
	_, err := s.pool.Exec(ctx, q)
	return err
}

func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s_%s", prefix, time.Now().UTC().Format("20060102150405"), hex.EncodeToString(b))
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
		_, err := s.pool.Exec(ctx, `insert into users (username, email, dni, full_name, password_hash, role) values ($1,$2,$3,$4,$5,$6)`, u.Username, u.Email, u.DNI, u.FullName, u.PasswordHash, u.Role)
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
	_, err = s.pool.Exec(ctx, `insert into users (username, email, dni, full_name, password_hash, role) values ($1,$2,$3,$4,$5,$6)`, user.Username, user.Email, user.DNI, user.FullName, user.PasswordHash, user.Role)
	return err
}

func (s *Store) GetByUsername(ctx context.Context, username string) (domain.User, bool, error) {
	var u domain.User
	err := s.pool.QueryRow(ctx, `select username, coalesce(email,''), coalesce(dni,''), coalesce(full_name,''), password_hash, role from users where username=$1`, username).Scan(&u.Username, &u.Email, &u.DNI, &u.FullName, &u.PasswordHash, &u.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, false, nil
		}
		return domain.User{}, false, err
	}
	return u, true, nil
}

func (s *Store) GetOrCreateByNombre(ctx context.Context, nombre string) (domain.Cliente, error) {
	nombreKey := util.NormalizeKey(nombre)
	var c domain.Cliente
	err := s.pool.QueryRow(ctx, `select id,nombre,nombre_key,creado_en from clientes where nombre_key=$1`, nombreKey).Scan(&c.ID, &c.Nombre, &c.NombreKey, &c.CreadoEn)
	if err == nil {
		return c, nil
	}
	if err != pgx.ErrNoRows {
		return domain.Cliente{}, err
	}
	c = domain.Cliente{ID: newID("cli"), Nombre: nombre, NombreKey: nombreKey, CreadoEn: time.Now().UTC()}
	_, err = s.pool.Exec(ctx, `insert into clientes (id,nombre,nombre_key,creado_en) values ($1,$2,$3,$4) on conflict (nombre_key) do nothing`, c.ID, c.Nombre, c.NombreKey, c.CreadoEn)
	if err != nil {
		return domain.Cliente{}, err
	}
	err = s.pool.QueryRow(ctx, `select id,nombre,nombre_key,creado_en from clientes where nombre_key=$1`, nombreKey).Scan(&c.ID, &c.Nombre, &c.NombreKey, &c.CreadoEn)
	return c, err
}

func (s *Store) GetByNombre(ctx context.Context, nombre string) ([]domain.Cliente, error) {
	key := util.NormalizeKey(nombre)
	out := make([]domain.Cliente, 0)
	var rows pgx.Rows
	var err error
	if key == "" {
		rows, err = s.pool.Query(ctx, `select id,nombre,nombre_key,creado_en from clientes order by creado_en asc`)
	} else {
		rows, err = s.pool.Query(ctx, `select id,nombre,nombre_key,creado_en from clientes where nombre_key=$1 order by creado_en asc`, key)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var c domain.Cliente
		if err := rows.Scan(&c.ID, &c.Nombre, &c.NombreKey, &c.CreadoEn); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) Create(ctx context.Context, sim domain.Simulacion) (domain.Simulacion, error) {
	sim.ID = newID("sim")
	sim.CreadoEn = time.Now().UTC()
	inJSON, err := json.Marshal(sim.Input)
	if err != nil {
		return domain.Simulacion{}, err
	}
	resJSON, err := json.Marshal(sim.Result)
	if err != nil {
		return domain.Simulacion{}, err
	}
	_, err = s.pool.Exec(ctx, `insert into simulaciones (id,cliente_id,creado_en,input_json,result_json) values ($1,$2,$3,$4,$5)`, sim.ID, sim.ClienteID, sim.CreadoEn, inJSON, resJSON)
	if err != nil {
		return domain.Simulacion{}, err
	}
	return sim, nil
}

func (s *Store) GetByID(ctx context.Context, id string) (domain.Simulacion, bool, error) {
	var sim domain.Simulacion
	var inJSON, resJSON []byte
	err := s.pool.QueryRow(ctx, `select id,cliente_id,creado_en,input_json,result_json from simulaciones where id=$1`, id).Scan(&sim.ID, &sim.ClienteID, &sim.CreadoEn, &inJSON, &resJSON)
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

func (s *Store) ListByClienteID(ctx context.Context, clienteID string) ([]domain.Simulacion, error) {
	rows, err := s.pool.Query(ctx, `select id,cliente_id,creado_en,input_json,result_json from simulaciones where cliente_id=$1 order by creado_en desc`, clienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Simulacion, 0)
	for rows.Next() {
		var sim domain.Simulacion
		var inJSON, resJSON []byte
		if err := rows.Scan(&sim.ID, &sim.ClienteID, &sim.CreadoEn, &inJSON, &resJSON); err != nil {
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
