package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	appcart "app/internal/application/cart"
	domaincart "app/internal/domain/cart"
	appmiddleware "app/internal/interfaces/http/middleware"
)

// MergeGuestCartIfNeeded merges a guest cart into the authenticated user's cart when applicable.
func MergeGuestCartIfNeeded(ctx context.Context, r *http.Request, carts *appcart.Service, owner domaincart.Owner) error {
	if carts == nil || owner.UserID == nil {
		return nil
	}
	token := appmiddleware.GetCartGuestToken(ctx)
	if token == "" {
		token = cartTokenFromRequest(r)
	}
	return carts.MergeGuestIntoUser(ctx, token, *owner.UserID)
}

func cartTokenFromRequest(r *http.Request) string {
	if token := r.Header.Get("X-Cart-Token"); token != "" {
		return token
	}
	cookie, err := r.Cookie(appmiddleware.CartTokenCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// OwnerWithUser returns a cart owner bound to the given user while preserving the guest token.
func OwnerWithUser(r *http.Request, userID uuid.UUID) domaincart.Owner {
	owner := domaincart.Owner{UserID: &userID}
	if token := appmiddleware.GetCartGuestToken(r.Context()); token != "" {
		owner.GuestToken = token
	} else if token := cartTokenFromRequest(r); token != "" {
		owner.GuestToken = token
	}
	return owner
}
