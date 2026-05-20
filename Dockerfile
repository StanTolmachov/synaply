# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install dependencies separately for caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app ./cmd/api

# Stage 2: Run
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/migrations ./migrations

HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080

CMD ["./app"]