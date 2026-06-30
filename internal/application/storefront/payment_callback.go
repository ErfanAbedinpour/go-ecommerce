package storefront

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"

	apporder "app/internal/application/order"
	"app/pkg/apperror"
)

// PaymentCallbackInput holds PSP callback data.
type PaymentCallbackInput struct {
	OrderID   uuid.UUID
	Authority string
	Status    string
	Signature string
}

// PaymentCallbackOutput is the payment callback response.
type PaymentCallbackOutput struct {
	OrderID       uuid.UUID `json:"order_id"`
	PaymentStatus string    `json:"payment_status"`
}

// HandlePaymentCallback verifies the PSP callback and updates order payment state.
func (s *Service) HandlePaymentCallback(ctx context.Context, input PaymentCallbackInput, callbackSecret string) (*PaymentCallbackOutput, error) {
	if !verifyPaymentSignature(callbackSecret, input.OrderID.String(), input.Authority, input.Status, input.Signature) {
		return nil, apperror.Unauthorized("invalid payment callback signature")
	}

	status := strings.ToUpper(strings.TrimSpace(input.Status))
	if status != "OK" && status != "NOK" {
		return nil, apperror.Validation("invalid payment status", map[string]string{
			"status": "must be OK or NOK",
		})
	}

	if status == "NOK" {
		order, err := s.orders.GetByID(ctx, input.OrderID)
		if err != nil {
			return nil, err
		}
		return &PaymentCallbackOutput{
			OrderID:       order.ID,
			PaymentStatus: order.PaymentStatus.String(),
		}, nil
	}

	order, err := s.orders.ConfirmPayment(ctx, apporder.ConfirmPaymentInput{
		OrderID:       input.OrderID,
		TransactionID: strings.TrimSpace(input.Authority),
	})
	if err != nil {
		return nil, err
	}

	if s.mailer != nil && order.Customer != nil && order.Customer.Email != "" {
		go func() {
			_ = s.mailer.SendOrderConfirmation(context.Background(), order.Customer.Email, order.OrderNumber, order.Total)
		}()
	}

	return &PaymentCallbackOutput{
		OrderID:       order.ID,
		PaymentStatus: order.PaymentStatus.String(),
	}, nil
}

func verifyPaymentSignature(secret, orderID, authority, status, signature string) bool {
	if secret == "" {
		return true
	}
	if signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(orderID + authority + status))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
