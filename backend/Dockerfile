# ── Development stage ──────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS dev

RUN apk add --no-cache git curl
# Air for hot-reload in development
RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# In dev, source is mounted as a volume — Air watches for changes
CMD ["air", "-c", ".air.toml"]


# ── Builder stage ───────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api


# ── Production stage ────────────────────────────────────────────────────────────
FROM alpine:3.20 AS prod

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/api .

EXPOSE 8080

CMD ["./api"]
