# ===== Build stage =====
FROM golang:1.22-alpine AS builder
WORKDIR /src

# (Opcional) instalar build tools adicionais
RUN apk add --no-cache git

# Copie os arquivos de módulo primeiro (para cache mais eficiente)
COPY go.mod go.sum ./
RUN go mod download

# Copie o restante do código (quando existir)
COPY . .

# Compile o binário
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/usersvc ./cmd/api

# ===== Runtime stage =====
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/usersvc /app/usersvc
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/usersvc"]
```

#> **Nota:** O `go.mod`/`go.sum` e o código em `cmd/api` serão adicionados quando você iniciar a implementação do servidor Gin (Sprint 1). O Compose e o Swagger já funcionam agora para documentação.
