package main

import (
	"context"
	"github.com/hectorgimenez/d2go/pkg/memory"
	zapLogger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/helper/winproc"
	asm "github.com/hectorgimenez/koolo/internal/memory"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/remote/discord"
	"github.com/hectorgimenez/koolo/internal/remote/telegram"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger, err := zapLogger.NewLogger(config.Config.Debug.Log, config.Config.LogFilePath)
	if err != nil {
		log.Fatalf("Error starting logger: %s", err.Error())
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ch
		logger.Info("Shutting down...")
		signal.Stop(ch)
		cancel()
	}()
	g, ctx := errgroup.WithContext(ctx)

	process, err := memory.NewProcess()
	if err != nil {
		logger.Fatal("Error finding D2R.exe process", zap.Error(err))
	}
	err = asm.ASMInjectorInit(uint32(process.GetPID()))
	if err != nil {
		logger.Fatal("Error initializing ASMInjector", zap.Error(err))
	}
	defer asm.ASMInjectorUnload()

	// Prevent screen from turning off
	winproc.SetThreadExecutionState.Call(winproc.EXECUTION_STATE_ES_DISPLAY_REQUIRED | winproc.EXECUTION_STATE_ES_CONTINUOUS)

	gr := &reader.GameReader{
		GameReader: memory.NewGameReader(process),
	}

	bm := health.NewBeltManager(logger)
	gm := helper.NewGameManager(gr)
	hm := health.NewHealthManager(logger, bm, gm)
	sm := town.NewShopManager(logger, bm)
	char, err := character.BuildCharacter(logger)
	if err != nil {
		logger.Fatal("Error creating character", zap.Error(err))
	}

	ab := action.NewBuilder(logger, sm, bm, gr, char)
	bot := koolo.NewBot(logger, hm, ab, gr)

	var supervisor koolo.Supervisor
	var companion koolo.Companion
	if config.Config.Companion.Enabled {
		supervisor = koolo.NewCompanionSupervisor(logger, bot, gr, gm)
		companion = supervisor.(koolo.Companion)
	} else {
		supervisor = koolo.NewSinglePlayerSupervisor(logger, bot, gr, gm)
	}

	g.Go(func() error {
		return supervisor.Start(ctx, run.NewFactory(logger, ab, char, gr, bm))
	})

	eventListener := event.NewListener(logger)

	// Discord Bot initialization
	if config.Config.Discord.Enabled {
		discordBot, err := discord.NewBot(config.Config.Discord.Token, config.Config.Discord.ChannelID, supervisor, companion)
		if err != nil {
			logger.Fatal("Discord could not been initialized", zap.Error(err))
		}
		eventListener.Register(discordBot.Handle)

		g.Go(func() error {
			return discordBot.Start(ctx)
		})
	}

	// Telegram Bot initialization
	if config.Config.Telegram.Enabled {
		telegramBot, err := telegram.NewBot(config.Config.Telegram.Token, config.Config.Telegram.ChatID, logger, supervisor)
		if err != nil {
			logger.Fatal("Telegram could not been initialized", zap.Error(err))
		}
		eventListener.Register(telegramBot.Handle)

		g.Go(func() error {
			return telegramBot.Start(ctx)
		})
	}

	g.Go(func() error {
		return eventListener.Listen(ctx)
	})

	err = g.Wait()
	if err != nil {
		logger.Fatal("Error running Koolo", zap.Error(err))
	}
}
