package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	domaincustomer "app/internal/domain/customer"
	"app/internal/domain/user"
)

func (s *AuthService) linkOrCreateCustomer(ctx context.Context, u *user.User) error {
	if u.Role != user.RoleCustomer {
		return nil
	}

	if _, err := s.customers.FindByUserID(ctx, u.ID); err == nil {
		return nil
	} else if err != domaincustomer.ErrNotFound {
		return err
	}

	var guest *domaincustomer.Customer
	if u.Email != "" {
		if g, err := s.customers.FindGuestByEmail(ctx, u.Email); err == nil {
			guest = g
		} else if err != domaincustomer.ErrNotFound {
			return err
		}
	}
	if guest == nil && u.Phone != "" {
		if g, err := s.customers.FindGuestByPhone(ctx, u.Phone); err == nil {
			guest = g
		} else if err != domaincustomer.ErrNotFound {
			return err
		}
	}

	now := time.Now().UTC()
	if guest != nil {
		guest.UserID = &u.ID
		guest.Type = domaincustomer.TypeRegistered
		guest.UpdatedAt = now
		return s.customers.Update(ctx, guest)
	}

	userID := u.ID
	customer := &domaincustomer.Customer{
		ID:        uuid.New(),
		UserID:    &userID,
		Type:      domaincustomer.TypeRegistered,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.customers.Create(ctx, customer)
}

// SessionOutput holds authentication tokens and the authenticated user id.
type SessionOutput struct {
	*TokenOutput
	UserID uuid.UUID `json:"-"`
}
