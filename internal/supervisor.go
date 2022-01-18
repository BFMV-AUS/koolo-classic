package koolo

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/config"
)

// Supervisor is the main bot entrypoint, it will handle all the parallel processes and ensure everything is up and running
type Supervisor struct {
	cfg           config.Config
	healthManager HealthManager
	bot           Bot
}

func NewSupervisor(cfg config.Config, hm HealthManager, bot Bot) Supervisor {
	return Supervisor{
		cfg:           cfg,
		healthManager: hm,
		bot:           bot,
	}
}

// Start will stay running during the application lifecycle, it will orchestrate all the required bot pieces
func (s Supervisor) Start(ctx context.Context) error {
	s.bot.Start(ctx)
	s.healthManager.Start(ctx)

	return nil
}
