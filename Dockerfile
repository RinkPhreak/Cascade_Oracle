# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies for build and cgo (if needed, though we avoid cgo where possible)
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the monolith
# CGO_ENABLED=0 ensures a fully static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o cascade ./cmd/server

# Run stage
FROM alpine:3.19

WORKDIR /app

# Import CA certs and tzdata from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Europe/Moscow

COPY --from=builder /app/cascade /app/cascade

# Since we use a monolithic entrypoint, we can control modes via args or env
# Defaults to running the monolith (API + Workers)
ENTRYPOINT ["/app/cascade"]
