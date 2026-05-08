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

## Docker
Build de imagen:
```bash
docker build -t backend-of .
```

Ejecutar contenedor:
```bash
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e JWT_SECRET=change-this-secret \
  -e DATABASE_URL="postgresql://usuario:password@host:5432/postgres" \
  backend-of
```

Si no usaras PostgreSQL, puedes omitir `DATABASE_URL` y usar:
```bash
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e JWT_SECRET=change-this-secret \
  -e DATA_FILE=/app/data.json \
  backend-of
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

POST /api/v1/vehiculos
GET  /api/v1/vehiculos
GET  /api/v1/vehiculos/:id

POST /api/v1/simulaciones
GET  /api/v1/simulaciones
GET  /api/v1/simulaciones/:id

GET  /api/v1/clientes/me

GET  /openapi.json
GET  /health
```

`GET /api/v1/simulaciones/:id` solo devuelve simulaciones del usuario autenticado.
`GET /api/v1/vehiculos/:id` aplica la misma regla de seguridad.

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

## Registrar vehiculo

```json
{
  "marca": "Toyota",
  "modelo": "Yaris",
  "anio": 2025,
  "tipo": "sedan",
  "precio": 80000,
  "moneda": "PEN"
}
```

## Historial filtrable

```text
GET /api/v1/simulaciones?moneda=PEN&plazoMeses=36&vehiculo=Yaris&montoMin=50000&montoMax=90000
```

Filtros disponibles:

- `fechaDesde` y `fechaHasta` con formato `YYYY-MM-DD`.
- `moneda`: `PEN` o `USD`.
- `plazoMeses`: `24` o `36`.
- `montoMin` y `montoMax`.
- `vehiculo`: busca por marca, modelo o tipo.

## Documentacion tecnica

El backend publica una especificacion OpenAPI basica en:

```text
GET /openapi.json
```

Esto permite mostrar en el informe los endpoints, modelos principales y parametros del historial sin depender de capturas manuales.

## Mejoras frente a una calculadora simple

- Login y registro con datos de identidad: username, email, DNI y nombre completo.
- Catalogo de vehiculos por usuario.
- Simulador con campos propios de Compra Inteligente: tasa nominal/efectiva, capitalizacion, plazo, cuota balloon, seguros, moneda y gracia.
- Cronograma con periodo, fecha, saldo inicial, cuota, interes, seguros, amortizacion y saldo final.
- Resumen financiero listo para UI: monto financiado, cuota inicial, cuota mensual, balloon, intereses, seguros, total pagado, TCEA, VAN y TIR.
- Historial filtrable de simulaciones.
- Seguridad por propietario en simulaciones y vehiculos.
