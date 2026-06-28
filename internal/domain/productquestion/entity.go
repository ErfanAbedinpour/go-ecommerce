package productquestion

import (
	"time"

	"github.com/google/uuid"
)

// Status tracks whether a question has been answered.
type Status string

const (
	StatusOpen     Status = "open"
	StatusAnswered Status = "answered"
)

// Question represents a customer question about a product.
type Question struct {
	ID         uuid.UUID
	ProductID  uuid.UUID
	AskerName  string
	AskerEmail string
	Question   string
	Answer     string
	AnsweredAt *time.Time
	AnsweredBy *uuid.UUID
	Status     Status
	CreatedAt  time.Time
}

// AdminListItem includes product name for admin views.
type AdminListItem struct {
	Question
	ProductName string
}
