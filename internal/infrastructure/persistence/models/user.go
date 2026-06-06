package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserModel is the GORM model for admin_users table.
type UserModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string         `gorm:"type:varchar(255);not null"`
	FirstName    string         `gorm:"type:varchar(100);not null"`
	LastName     string         `gorm:"type:varchar(100);not null"`
	Phone        *string        `gorm:"type:varchar(20)"`
	Role         string         `gorm:"type:varchar(20);not null;default:customer"`
	IsActive     bool           `gorm:"not null;default:true"`
	LastLoginAt  *time.Time     `gorm:"type:timestamptz"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;not null;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string { return "admin_users" }

// RefreshTokenModel is the GORM model for refresh_tokens table.
type RefreshTokenModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdminUserID uuid.UUID  `gorm:"type:uuid;not null;index;column:admin_user_id"`
	TokenHash   string     `gorm:"type:varchar(255);not null;index"`
	FamilyID    uuid.UUID  `gorm:"type:uuid;not null"`
	ExpiresAt   time.Time  `gorm:"type:timestamptz;not null"`
	RevokedAt   *time.Time `gorm:"type:timestamptz"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }
