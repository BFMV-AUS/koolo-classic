package koolo

import (
	"context"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/lxn/win"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"time"
)

// Supervisor is the main bot entrypoint, it will handle all the parallel processes and ensure everything is up and running
type Supervisor struct {
	logger        *zap.Logger
	cfg           config.Config
	eventsChannel <-chan event.Event
	ah            action.Handler
	healthManager health.Manager
	bot           game.Bot
}

func NewSupervisor(logger *zap.Logger, cfg config.Config, ah action.Handler, hm health.Manager, bot game.Bot) Supervisor {
	return Supervisor{
		logger:        logger,
		cfg:           cfg,
		ah:            ah,
		healthManager: hm,
		bot:           bot,
	}
}

// Start will stay running during the application lifecycle, it will orchestrate all the required bot pieces
func (s Supervisor) Start(ctx context.Context) error {
	err := s.ensureProcessIsRunningAndPrepare()
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	// Listen to actions triggered from elsewhere
	g.Go(func() error {
		return s.ah.Listen(ctx)
	})

	// Listen to events and attaching/detaching bot operations
	g.Go(func() error {
		s.listenEvents(ctx)
		return nil
	})

	// Main loop will be inside this, will handle bosses and path traveling
	g.Go(func() error {
		return s.bot.Start(ctx)
	})

	// Will keep our character and mercenary alive, monitoring life and mana
	g.Go(func() error {
		return s.healthManager.Start(ctx)
	})

	return g.Wait()
}

func (s Supervisor) listenEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-s.eventsChannel:
			switch evt {
			case event.ExitedGame:
				s.healthManager.Pause()
			case event.SafeAreaAbandoned:
				s.healthManager.Resume()
			case event.SafeAreaEntered:
				s.healthManager.Pause()
			}
		}
	}
}

func (s Supervisor) ensureProcessIsRunningAndPrepare() error {
	window := robotgo.FindWindow("Diablo II: Resurrected")
	if window == win.HWND_TOP {
		s.logger.Fatal("Diablo II: Resurrected window can not be found! Are you sure game is open?")
	}
	win.SetForegroundWindow(window)

	// Exclude border offsets
	// TODO: Improve this, maybe getting window content coordinates?
	pos := win.WINDOWPLACEMENT{}
	win.GetWindowPlacement(window, &pos)
	hid.WindowLeftX = int(pos.RcNormalPosition.Left) + 8
	hid.WindowTopY = int(pos.RcNormalPosition.Top) + 31
	hid.GameAreaSizeX = int(pos.RcNormalPosition.Right) - hid.WindowLeftX - 10
	hid.GameAreaSizeY = int(pos.RcNormalPosition.Bottom) - hid.WindowTopY - 10
	time.Sleep(time.Second * 1)

	return nil
}
