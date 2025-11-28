# ===== Build stage =====
FROM golang:1.22-alpine AS builder
ARG USE_PREBUILT=0
WORKDIR /src

# (Opcional) instalar build tools adicionais
RUN apk add --no-cache git

# Copie os arquivos de módulo primeiro (para cache mais eficiente)
COPY go.mod go.sum ./
RUN go mod download

# Copie o restante do código
COPY . .

RUN mkdir -p /out

# Se houver um binário pré-compilado (cmd/api/usersvc) e USE_PREBUILT=1, reutiliza.
# Caso contrário, compila a partir do código fonte.
RUN if [ "$USE_PREBUILT" = "1" ] && [ -f "/src/cmd/api/usersvc" ]; then \
      echo "Using prebuilt binary"; \
      cp /src/cmd/api/usersvc /out/usersvc; \
    else \
      echo "Building binary"; \
      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/usersvc ./cmd/api; \
    fi

# ===== Runtime stage =====
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/usersvc /app/usersvc
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/usersvc"]
