package models

import (
	"time"

	"github.com/google/uuid"
)

// ContactMessageModel is the GORM model for contact_messages table.
type ContactMessageModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Email     string    `gorm:"type:varchar(255);not null"`
	Phone     *string   `gorm:"type:varchar(50)"`
	Subject   *string   `gorm:"type:varchar(500)"`
	Message   string    `gorm:"type:text;not null"`
	Source    string    `gorm:"type:contact_message_source;not null;default:'homepage'"`
	Status   string    `gorm:"type:contact_message_status;not null;default:'unread'"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (ContactMessageModel) TableName() string { return "contact_messages" }

// WishlistItemModel is the GORM model for wishlist_items table.
type WishlistItemModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (WishlistItemModel) TableName() string { return "wishlist_items" }

// ProductReviewModel is the GORM model for product_reviews table.
type ProductReviewModel struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProductID  uuid.UUID  `gorm:"type:uuid;not null"`
	CustomerID *uuid.UUID `gorm:"type:uuid"`
	AuthorName string     `gorm:"type:varchar(255);not null"`
	Rating     int        `gorm:"type:smallint;not null"`
	Title      *string    `gorm:"type:varchar(255)"`
	Content    string     `gorm:"type:text;not null"`
	Status     string     `gorm:"type:product_review_status;not null;default:'pending'"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (ProductReviewModel) TableName() string { return "product_reviews" }

// ProductQuestionModel is the GORM model for product_questions table.
type ProductQuestionModel struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProductID  uuid.UUID  `gorm:"type:uuid;not null"`
	AskerName  string     `gorm:"type:varchar(255);not null"`
	AskerEmail *string    `gorm:"type:varchar(255)"`
	Question   string     `gorm:"type:text;not null"`
	Answer     *string    `gorm:"type:text"`
	AnsweredAt *time.Time `gorm:"type:timestamptz"`
	AnsweredBy *uuid.UUID `gorm:"type:uuid"`
	Status     string     `gorm:"type:product_question_status;not null;default:'open'"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (ProductQuestionModel) TableName() string { return "product_questions" }
