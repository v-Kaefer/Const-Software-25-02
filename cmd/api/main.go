package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/v-Kaefer/Const-Software-25-02/internal/config"
	appdb "github.com/v-Kaefer/Const-Software-25-02/internal/db"
	httpapi "github.com/v-Kaefer/Const-Software-25-02/internal/http"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/workspace"
)

func main() {

	fmt.Println("Hello, 世界")

	// 1) Config (env, DSN, env=development|production)
	cfg := config.Load()

	// 2) DB (GORM + pool + logger)
	gormDB, err := appdb.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// 3) Migração automática só em dev (para prototipagem)
	if cfg.Env != "production" {
		if err := appdb.AutoMigrate(gormDB); err != nil {
			log.Fatal(err)
		}
	}

	// 4) Repositórios e serviços (injeção de dependências)
	userRepo := user.NewRepo(gormDB)
	userSvc := user.NewService(gormDB, userRepo)
	projectSvc := workspace.NewProjectService(gormDB)
	taskSvc := workspace.NewTaskService(gormDB)
	timeSvc := workspace.NewTimeEntryService(gormDB)

	// 5) Auth middleware (configuração do Cognito)
	authMiddleware := httpapi.NewAuthMiddleware(cfg.Cognito)

	// 6) HTTP router (camada de entrega, não conhece GORM)
	router := httpapi.NewRouter(userSvc, projectSvc, taskSvc, timeSvc, authMiddleware)

	// 7) CORS middleware
	handler := corsMiddleware(router)

	// 8) Servidor + graceful shutdown
	port := getenv("APP_PORT", "8080")
	addr := ":" + port
	srv := &http.Server{Addr: addr, Handler: handler}

	go func() {
		log.Printf("listening on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)

	sqlDB, _ := gormDB.DB()
	_ = sqlDB.Close()
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins in development, restrict in production
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getenv returns the value of an environment variable or a default value
func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
