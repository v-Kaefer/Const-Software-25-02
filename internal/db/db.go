package db

import (
	"time"
	"Const-Software-25-02/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// Open abre a conexão GORM com Postgres e configura o pool.
func Open(cfg config.AppConfig) (*gorm.DB, error) {
	lvl := logger.Warn
	if cfg.Env == "development" {
		lvl = logger.Info
	}

	gormCfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // tabelas no plural
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:      logger.Default.LogMode(lvl),
		PrepareStmt: true,
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DB.DSN(),
		PreferSimpleProtocol: true,
	}), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(60 * time.Minute)

	return db, sqlDB.Ping()
}
