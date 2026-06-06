package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminUserModel is the GORM model for admin_users table.
type AdminUserModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string         `gorm:"type:varchar(255);not null"`
	FirstName    string         `gorm:"type:varchar(100);not null"`
	LastName     string         `gorm:"type:varchar(100);not null"`
	Phone        *string        `gorm:"type:varchar(20)"`
	IsActive     bool           `gorm:"not null;default:true"`
	LastLoginAt  *time.Time     `gorm:"type:timestamptz"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;not null;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Roles        []RoleModel    `gorm:"many2many:admin_user_roles;"`
}

func (AdminUserModel) TableName() string { return "admin_users" }

// RoleModel is the GORM model for roles table.
type RoleModel struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string            `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description *string           `gorm:"type:text"`
	CreatedAt   time.Time         `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"type:timestamptz;not null;autoUpdateTime"`
	Permissions []PermissionModel `gorm:"many2many:role_permissions;"`
}

func (RoleModel) TableName() string { return "roles" }

// PermissionModel is the GORM model for permissions table.
type PermissionModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Description *string   `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (PermissionModel) TableName() string { return "permissions" }

// RefreshTokenModel is the GORM model for refresh_tokens table.
type RefreshTokenModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdminUserID uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash   string     `gorm:"type:varchar(255);not null;index"`
	FamilyID    uuid.UUID  `gorm:"type:uuid;not null"`
	ExpiresAt   time.Time  `gorm:"type:timestamptz;not null"`
	RevokedAt   *time.Time `gorm:"type:timestamptz"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }
