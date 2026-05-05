# Backend OF - API de Simulador de Credito Vehicular

## Stack
- Go 1.26+
- Gin (HTTP)
- Persistencia en archivo JSON (`data.json`) con repositorios desacoplados
- Auth con cookies `httpOnly` firmadas (access + refresh)

> Nota: el entorno no permitio descargar drivers SQL. El codigo esta desacoplado por repositorio para cambiar a SQLite/PostgreSQL sin tocar handlers ni dominio.

## Estructura
- `main.go`: bootstrap
- `internal/config`: variables de entorno
- `internal/domain`: modelos y errores API
- `internal/auth`: token firmado + rate limit login
- `internal/repository`: interfaces
- `internal/repository/jsondb`: implementacion de persistencia actual
- `internal/service`: reglas de negocio (auth/simulacion)
- `internal/transport/http`: router, middleware y handlers

## Variables de entorno
- `PORT` (default `8080`)
- `JWT_SECRET` (default `change-this-secret`)
- `DATA_FILE` (default `./data.json`)
- `ACCESS_TTL_MIN` (default `15`)
- `REFRESH_TTL_HOURS` (default `24`)
- `COOKIE_SECURE` (default `false`)
- `LOGIN_MAX_PER_MIN` (default `10`)

## Ejecutar
```bash
go run .
```

## Tests
```bash
go test ./...
```

## Auth seeds
- `admin/admin`
- `user/user`

## API (v1)
Base: `/api/v1`

### POST `/auth/login`
Body:
```json
{ "username": "admin", "password": "admin" }
```
Response 200:
```json
{ "user": { "username": "admin" } }
```
Setea cookies: `access_token`, `refresh_token`.

### POST `/auth/logout`
Response 200:
```json
{ "ok": true }
```

### POST `/auth/refresh`
Response 200:
```json
{ "ok": true }
```
Rota cookies de access/refresh.

### GET `/auth/session` (protegido)
Response 200:
```json
{ "authenticated": true, "user": { "username": "admin" } }
```

### POST `/simulaciones` (protegido)
Body:
```json
{
  "nombreCliente": "Juan Perez",
  "precioVehiculo": 50000,
  "porcentajeCuotaInicial": 10,
  "plazoMeses": 24,
  "tasaNominalAnual": 12,
  "capitalizacionPorAnio": 12,
  "periodosGracia": 2,
  "tipoGracia": "parcial"
}
```
Response 201: `Simulacion` completa con cronograma, VAN, TIR.

### GET `/simulaciones?cliente=juan perez` (protegido)
Response 200:
```json
{ "items": [ ... ] }
```

### GET `/simulaciones/:id` (protegido)
Response 200: `Simulacion`
Response 404:
```json
{ "error": { "code": "not_found", "message": "simulacion no encontrada", "field": "id" } }
```

## Contratos de dominio
- `Pago`: `{ mes, cuota, interes, amortizacion, saldoDeudor }`
- `SimulacionInput`: `{ nombreCliente, precioVehiculo, porcentajeCuotaInicial, plazoMeses, tasaNominalAnual, capitalizacionPorAnio, periodosGracia, tipoGracia }`
- `SimulacionResult`: `{ tasaPeriodo, van, tir, cronograma: Pago[] }`
- `Simulacion`: `{ id, clienteId, creadoEn, input, result }`
- `Cliente`: `{ id, nombre, nombreKey, creadoEn }`

## Errores JSON
Formato:
```json
{
  "error": { "code": "validation_error", "message": "errores de validacion" },
  "errors": [
    { "code": "validation_error", "message": "plazoMeses debe ser >= 1", "field": "plazoMeses" }
  ]
}
```

## Regla de calculo implementada
- Capital: `precio - precio*(cuotaInicial/100)`
- Tasa efectiva mensual desde TN: `i = (1 + TN/m)^(m/12) - 1` (si TN > 1, se interpreta `%`)
- Gracia: `min(plazo, periodosGracia)`
- Frances para meses sin gracia
- Gracia total: capitaliza interes
- Gracia parcial: paga solo interes
- Ultima cuota ajustada para no sobre-amortizar
- VAN con `t=0` incluido y flujos banco (`-capital`, luego `+cuotas`)
- TIR por Newton-Raphson con fallback por biseccion

## Esquema SQL recomendado (migracion futura)
```sql
CREATE TABLE users (
  username TEXT PRIMARY KEY,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE clientes (
  id TEXT PRIMARY KEY,
  nombre TEXT NOT NULL,
  nombre_key TEXT NOT NULL UNIQUE,
  creado_en TIMESTAMP NOT NULL
);

CREATE TABLE simulaciones (
  id TEXT PRIMARY KEY,
  cliente_id TEXT NOT NULL REFERENCES clientes(id),
  creado_en TIMESTAMP NOT NULL,
  input_json JSONB NOT NULL,
  result_json JSONB NOT NULL
);

CREATE INDEX idx_sim_cliente_creado ON simulaciones(cliente_id, creado_en DESC);
```

## Cambios minimos en frontend Astro
1. Reemplazar login local por `POST /api/v1/auth/login` con `credentials: 'include'`.
2. Reemplazar guard de `/` por `GET /api/v1/auth/session`.
3. Reemplazar escritura Dexie de simulacion por `POST /api/v1/simulaciones`.
4. Reemplazar lecturas historicas por `GET /api/v1/simulaciones?cliente=...` y `GET /api/v1/simulaciones/:id`.
5. Agregar logout con `POST /api/v1/auth/logout`.
6. Ante `401`, llamar `POST /api/v1/auth/refresh` y reintentar.
