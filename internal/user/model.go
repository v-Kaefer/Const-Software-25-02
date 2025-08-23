package user

import "time"

// User representa um usuário do sistema.
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"uniqueIndex;size:255;not null"`
	Name      string    `gorm:"size:120;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
