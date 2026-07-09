# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ---- Test stage ----
FROM builder AS tester

RUN go vet ./...
RUN go test ./... -v

# ---- Compile Stage ----
FROM tester AS build

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/importer ./cmd/import
RUN CGO_ENABLED=0 GOOS=linux GOBIN=/app/bin go install github.com/pressly/goose/v3/cmd/goose@v3.27.1

# ---- Runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /app/bin/api ./api
COPY --from=build /app/bin/importer ./importer
COPY --from=build /app/bin/goose ./goose
COPY scripts/migrate-and-import.sh ./migrate-and-import.sh
COPY data ./data
COPY sql/schema ./sql/schema

# godotenv.Load() errors if no .env file exists in the working dir.
# Real config comes from env vars set by docker-compose, which
# godotenv will not override.
RUN touch .env

EXPOSE 8080

CMD ["./api"]
