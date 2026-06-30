package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	domaincart "app/internal/domain/cart"
)

type contextKeyCart string

const (
	CartOwnerKey      contextKeyCart = "cart_owner"
	CartGuestTokenKey contextKeyCart = "cart_guest_token"
	CartTokenCookie                  = "cart_token"
)

// CartSession resolves the guest cart token and authenticated owner for store cart routes.
func CartSession() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := cartTokenFromRequest(r)
			created := false
			if token == "" {
				token = uuid.New().String()
				setCartTokenCookie(w, token)
				created = true
			}

			owner := domaincart.Owner{GuestToken: token}
			if userID, ok := GetUserIDOptional(r.Context()); ok {
				owner.UserID = &userID
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, CartOwnerKey, owner)
			ctx = context.WithValue(ctx, CartGuestTokenKey, token)
			if created {
				ctx = context.WithValue(ctx, contextKeyCart("cart_token_created"), true)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetCartOwner returns the cart owner from request context.
func GetCartOwner(ctx context.Context) (domaincart.Owner, bool) {
	owner, ok := ctx.Value(CartOwnerKey).(domaincart.Owner)
	return owner, ok
}

// GetCartGuestToken returns the guest cart token from request context.
func GetCartGuestToken(ctx context.Context) string {
	token, _ := ctx.Value(CartGuestTokenKey).(string)
	return token
}

func cartTokenFromRequest(r *http.Request) string {
	if token := r.Header.Get("X-Cart-Token"); token != "" {
		return token
	}
	cookie, err := r.Cookie(CartTokenCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setCartTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CartTokenCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
