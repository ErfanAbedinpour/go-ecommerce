package order

import "testing"

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from, to Status
		want     bool
	}{
		{StatusPending, StatusProcessing, true},
		{StatusPending, StatusCancelled, true},
		{StatusProcessing, StatusShipped, true},
		{StatusShipped, StatusDelivered, true},
		{StatusDelivered, StatusRefunded, true},
		{StatusShipped, StatusCancelled, false},
		{StatusCancelled, StatusProcessing, false},
		{StatusRefunded, StatusDelivered, false},
	}

	for _, tt := range tests {
		got := CanTransition(tt.from, tt.to)
		if got != tt.want {
			t.Errorf("CanTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestCanCancel(t *testing.T) {
	if !CanCancel(StatusPending) {
		t.Error("pending should be cancellable")
	}
	if CanCancel(StatusShipped) {
		t.Error("shipped should not be cancellable")
	}
}

func TestCanRefund(t *testing.T) {
	if !CanRefund(StatusDelivered, PaymentPaid) {
		t.Error("delivered+paid should be refundable")
	}
	if CanRefund(StatusDelivered, PaymentUnpaid) {
		t.Error("delivered+unpaid should not be refundable")
	}
}
