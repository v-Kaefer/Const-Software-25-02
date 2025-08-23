# ===== Build stage =====
FROM golang:1.22-alpine AS builder
WORKDIR /src

# Optional: install git (used by go mod for private modules)
RUN apk add --no-cache git

# Go module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/usersvc ./cmd/api

# ===== Runtime stage =====
FROM gcr.io/distroless/base-debian12
WORKDIR /app

# Copy the binary
COPY --from=builder /out/usersvc /app/usersvc

# The service listens on 8080
EXPOSE 8080

# Run as non-root (provided by distroless)
USER nonroot:nonroot

# Start the service
ENTRYPOINT ["/app/usersvc"]
