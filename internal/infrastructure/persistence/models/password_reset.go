package models

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetTokenModel is the GORM model for password_reset_tokens table.
type PasswordResetTokenModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index;column:user_id"`
	TokenHash   string     `gorm:"type:varchar(255);not null;index"`
	ExpiresAt   time.Time  `gorm:"type:timestamptz;not null"`
	UsedAt      *time.Time `gorm:"type:timestamptz"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (PasswordResetTokenModel) TableName() string { return "password_reset_tokens" }
