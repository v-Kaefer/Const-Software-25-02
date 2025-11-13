# ===== Build stage =====
FROM golang:1.22 AS builder
WORKDIR /src

# Copie os arquivos de módulo e vendor directory
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copie o restante do código (quando existir)
COPY . .

# Compile o binário usando vendor mode
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /out/usersvc ./cmd/api

# ===== Runtime stage =====
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/usersvc /app/usersvc
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/usersvc"]

#> **Nota:** O `go.mod`/`go.sum` e o código em `cmd/api` serão adicionados quando você iniciar a implementação do servidor Gin (Sprint 1). O Compose e o Swagger já funcionam agora para documentação.
