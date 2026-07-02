package order

import (
	"context"
	"log/slog"
	"time"
)

const defaultExpiryInterval = 15 * time.Minute

// StartExpiryWorker runs a background loop that cancels unpaid orders past their payment window.
func StartExpiryWorker(ctx context.Context, svc *Service, interval time.Duration, log *slog.Logger) {
	if interval <= 0 {
		interval = defaultExpiryInterval
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		runExpirySweep(ctx, svc, log)

		for {
			select {
			case <-ctx.Done():
				log.Info("order expiry worker stopped")
				return
			case <-ticker.C:
				runExpirySweep(ctx, svc, log)
			}
		}
	}()
}

func runExpirySweep(ctx context.Context, svc *Service, log *slog.Logger) {
	expired, err := svc.ExpireUnpaidOrders(ctx, 50)
	if err != nil {
		log.Error("order expiry sweep failed", slog.String("error", err.Error()))
		return
	}
	if expired > 0 {
		log.Info("expired unpaid orders", slog.Int("count", expired))
	}
}
