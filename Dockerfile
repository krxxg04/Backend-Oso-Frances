FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo -ldflags="-s -w" -o /out/app ./cmd/backend

FROM alpine:3.22

RUN adduser -D -g '' appuser
WORKDIR /app

COPY --from=builder /out/app /app/app
COPY --from=builder /app/data.json /app/data.json

ENV PORT=8080
EXPOSE 8080

USER appuser

CMD ["/app/app"]
