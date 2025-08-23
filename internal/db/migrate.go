package db

import (
	"gorm.io/gorm"
	"Const-Software-25-02/internal/user"
)

// AutoMigrate roda as migrações a partir dos models.
// Use SOMENTE em dev; em produção prefira arquivos SQL versionados.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&user.User{},
	)
}
