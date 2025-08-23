package main

import (
	"fmt"
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"Const-Software-25-02/internal/config"
	appdb "Const-Software-25-02/internal/db"
	httpapi "Const-Software-25-02/internal/http"
	"Const-Software-25-02/internal/user"
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

	// 5) HTTP router (camada de entrega, não conhece GORM)
	router := httpapi.NewRouter(userSvc)

	// 6) Servidor + graceful shutdown
	srv := &http.Server{Addr: ":8080", Handler: router}

	go func() {
		log.Println("listening on :8080")
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
