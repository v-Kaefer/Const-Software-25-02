package user

import "time"

// User representa um usuário do sistema.
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"uniqueIndex;size:255;not null;uniqueIndex:idx_users_email"`
	Name      string    `gorm:"size:120;not null"`
	Role      string    `gorm:"size:50;default:'user-group'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
