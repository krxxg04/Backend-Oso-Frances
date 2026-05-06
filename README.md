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
