# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/importer ./cmd/import

# ---- Runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/api ./api
COPY --from=builder /app/bin/importer ./importer
COPY data ./data

# godotenv.Load() errors if no .env file exists in the working dir.
# Real config comes from env vars set by docker-compose, which
# godotenv will not override.
RUN touch .env

EXPOSE 8080

CMD ["./api"]
