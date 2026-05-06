# Backend OF - API de Simulador de Credito Vehicular

## Stack
- Go 1.26+
- Gin (HTTP)
- PostgreSQL (Supabase) via `DATABASE_URL`
- Fallback local JSON (`data.json`) si no defines `DATABASE_URL`
- Auth con cookies `httpOnly` firmadas (access + refresh)

## Variables de entorno
- `PORT` (default `8080`)
- `JWT_SECRET` (default `change-this-secret`)
- `DATABASE_URL` (si existe, usa PostgreSQL)
- `DATA_FILE` (default `./data.json`, solo fallback local)
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

## Nota DB
Al iniciar con `DATABASE_URL`, el backend ejecuta auto-migracion de:
- `users`
- `clientes`
- `simulaciones`

Y hace seed inicial de usuarios:
- `admin/admin`
- `user/user`

## Modelo funcional
El backend ahora representa mejor el producto Credito Vehicular "Compra Inteligente":

- `User`: acceso, email, DNI, nombre completo y rol.
- `Client`: datos del cliente vinculado a sus simulaciones.
- `Vehicle`: marca, modelo, anio, tipo, precio y moneda.
- `CreditSimulation`: entrada del credito, resultado financiero y cronograma.
- `PaymentSchedule`: periodo, fecha, saldo inicial, cuota, interes, seguros, amortizacion y saldo final.
- `Rate`: tasa efectiva o nominal, frecuencia de capitalizacion y tasa efectiva anual calculada.
- `Insurance`: seguro vehicular mensual y seguro de desgravamen anual.

## Endpoints

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
GET  /api/v1/auth/session

POST /api/v1/simulaciones
GET  /api/v1/simulaciones
GET  /api/v1/simulaciones/:id

GET  /health
```

`GET /api/v1/simulaciones/:id` solo devuelve simulaciones del usuario autenticado.

## Registro

```json
{
  "username": "cliente01",
  "email": "cliente01@email.com",
  "dni": "12345678",
  "fullName": "Cliente Demo",
  "password": "secret123"
}
```

## Crear simulacion

```json
{
  "moneda": "PEN",
  "vehiculo": {
    "marca": "Toyota",
    "modelo": "Yaris",
    "anio": 2025,
    "tipo": "sedan",
    "precio": 80000
  },
  "porcentajeCuotaInicial": 20,
  "plazoMeses": 36,
  "tipoTasa": "nominal",
  "tasaAnual": 18,
  "frecuenciaCapitalizacion": 12,
  "periodosPorAnio": 12,
  "cuotaFinalBalloon": 24000,
  "seguroVehicularMensual": 180,
  "seguroDesgravamenAnual": 1.2,
  "periodosGracia": 0,
  "tipoGracia": "sin_gracia",
  "fechaInicio": "2026-06-01"
}
```

La respuesta incluye un `resumen` con monto financiado, cuota inicial, cuota mensual, cuota final balloon, total de intereses, total de seguros, total pagado, TCEA, VAN, TIR y fecha de finalizacion. Tambien incluye un `cronograma` con saldos, intereses, seguros y amortizacion por periodo.
