package contact

import (
	"time"

	"github.com/google/uuid"
)

// Source indicates which form the message came from.
type Source string

const (
	SourceHomepage    Source = "homepage"
	SourceAbout       Source = "about"
	SourceContactPage Source = "contact_page"
)

// Status tracks admin inbox workflow.
type Status string

const (
	StatusUnread   Status = "unread"
	StatusRead     Status = "read"
	StatusArchived Status = "archived"
)

// InboxStats holds aggregate contact inbox counters.
type InboxStats struct {
	UnreadCount int64
	TotalCount  int64
}

// Message represents a contact form submission.
type Message struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Phone     string
	Subject   string
	Message   string
	Source    Source
	Status   Status
	CreatedAt time.Time
}
