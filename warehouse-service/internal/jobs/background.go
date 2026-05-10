// Package jobs фоновые задачи склада.
package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/industrial-sed/warehouse-service/internal/repositories"
	"github.com/industrial-sed/warehouse-service/internal/usecases"
)

// RunReservationExpirer периодически снимает просроченные резервы.
func RunReservationExpirer(ctx context.Context, log *slog.Logger, uc *usecases.UC, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n, err := uc.ProcessExpiredReservations(context.Background(), time.Now().UTC())
			if err != nil {
				log.Error("reservation_expirer", "error", err.Error())
				continue
			}
			if n > 0 {
				log.Info("reservation_expirer", "expired", n)
			}
		}
	}
}

// RunExpiryAlerts логирует партии с истекающим сроком (без рассылки).
func RunExpiryAlerts(ctx context.Context, log *slog.Logger, store *repositories.Store, every time.Duration, horizonDays int) {
	if horizonDays <= 0 {
		horizonDays = 30
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			until := time.Now().UTC().AddDate(0, 0, horizonDays)
			rows, err := store.ListExpiringBatches(context.Background(), nil, "", until)
			if err != nil {
				log.Error("expiry_alerts", "error", err.Error())
				continue
			}
			if len(rows) > 0 {
				log.Info("expiry_alerts", "batches", len(rows), "until", until.Format("2006-01-02"))
			}
		}
	}
}
