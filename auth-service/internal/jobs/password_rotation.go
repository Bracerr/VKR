package jobs

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Nerzal/gocloak/v13"
)

const requiredActionUpdatePassword = "UPDATE_PASSWORD"

type adminLogin interface {
	LoginAdmin(ctx context.Context) (*gocloak.JWT, error)
}

type userAdmin interface {
	GetUserByID(ctx context.Context, token, userID string) (*gocloak.User, error)
	UpdateUser(ctx context.Context, token string, user gocloak.User) error
}

type userIDSource interface {
	ListIDsCreatedBefore(ctx context.Context, before time.Time, limit int) ([]string, error)
}

type PasswordRotationConfig struct {
	Enabled bool
	After   time.Duration
	Every   time.Duration
	Batch   int
}

// RunPasswordRotation periodically enforces UPDATE_PASSWORD required action for old users.
func RunPasswordRotation(ctx context.Context, log *slog.Logger, cfg PasswordRotationConfig, kc adminLogin, users userAdmin, src userIDSource) {
	if log == nil {
		log = slog.Default()
	}
	if !cfg.Enabled {
		log.Info("password_rotation_disabled")
		return
	}
	if cfg.Every <= 0 {
		cfg.Every = 10 * time.Minute
	}
	if cfg.After <= 0 {
		cfg.After = 7 * 24 * time.Hour
	}
	if cfg.Batch <= 0 {
		cfg.Batch = 200
	}

	log.Info("password_rotation_started", "after", cfg.After.String(), "every", cfg.Every.String(), "batch", cfg.Batch)
	t := time.NewTicker(cfg.Every)
	defer t.Stop()

	runOnce := func() {
		before := time.Now().UTC().Add(-cfg.After)
		ids, err := src.ListIDsCreatedBefore(ctx, before, cfg.Batch)
		if err != nil {
			log.Warn("password_rotation_list_failed", "error", err.Error())
			return
		}
		if len(ids) == 0 {
			return
		}
		jwt, err := kc.LoginAdmin(ctx)
		if err != nil || jwt == nil || jwt.AccessToken == "" {
			if err == nil {
				err = errors.New("empty admin token")
			}
			log.Warn("password_rotation_admin_login_failed", "error", err.Error())
			return
		}
		token := jwt.AccessToken

		changed := 0
		for _, id := range ids {
			u, err := users.GetUserByID(ctx, token, id)
			if err != nil || u == nil {
				continue
			}
			already := false
			if u.RequiredActions != nil {
				for _, a := range *u.RequiredActions {
					if a == requiredActionUpdatePassword {
						already = true
						break
					}
				}
			}
			if already {
				continue
			}
			ra := []string{requiredActionUpdatePassword}
			u.RequiredActions = &ra
			// гарантируем, что UpdateUser точно знает ID
			if u.ID == nil || *u.ID == "" {
				u.ID = gocloak.StringP(id)
			}
			if err := users.UpdateUser(ctx, token, *u); err == nil {
				changed++
			}
		}
		if changed > 0 {
			log.Info("password_rotation_enforced", "users", changed, "cutoff", before.Format(time.RFC3339))
		}
	}

	// первый прогон сразу после старта
	runOnce()

	for {
		select {
		case <-ctx.Done():
			log.Info("password_rotation_stopped")
			return
		case <-t.C:
			runOnce()
		}
	}
}
