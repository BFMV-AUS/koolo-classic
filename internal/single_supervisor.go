package koolo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/run"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

type SinglePlayerSupervisor struct {
	*baseSupervisor
}

func NewSinglePlayerSupervisor(name string, bot *Bot, runFactory *run.Factory, statsHandler *StatsHandler, c container.Container) (*SinglePlayerSupervisor, error) {
	bs, err := newBaseSupervisor(bot, runFactory, name, statsHandler, c)
	if err != nil {
		return nil, err
	}

	return &SinglePlayerSupervisor{
		baseSupervisor: bs,
	}, nil
}

// Start will return error if it can not be started, otherwise will always return nil
func (s *SinglePlayerSupervisor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel

	err := s.ensureProcessIsRunningAndPrepare(ctx)
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	firstRun := true
	var gameCounter int
	var alreadyJoined bool
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Failsafe for when the bot is paused
			if s.bot.paused {

				// Sleep for a second before checking again
				helper.Sleep(1000)
				continue
			}

			if firstRun {
				err = s.waitUntilCharacterSelectionScreen()
				if err != nil {
					return fmt.Errorf("error waiting for character selection screen: %w", err)
				}
			}
			if !s.c.Manager.InGame() {

				if s.c.CharacterCfg.Game.EnableLobbyGames {

					// Failsafe in case neither create or join have been selected
					if !s.c.CharacterCfg.Game.CreateOnlineGames && !s.c.CharacterCfg.Game.JoinOnlineGame {
						s.c.CharacterCfg.Game.CreateOnlineGames = true
					}

					if s.c.CharacterCfg.Game.CreateOnlineGames {
						// I know this will also increase it from 0 to 1 the first time, who does a run with 0 lol
						gameCounter++
						if _, err = s.c.Manager.CreateOnlineGame(gameCounter); err != nil {
							s.c.Logger.Error(fmt.Sprintf("Error creating Online Game: %s", err.Error()))
							continue
						}
					} else if s.c.CharacterCfg.Game.JoinOnlineGame {
						if !alreadyJoined {
							if err = s.c.Manager.JoinOnlineGame(s.c.CharacterCfg.Game.OnlineGameNameTemplate, s.c.CharacterCfg.Game.OnlineGamePassowrd); err != nil {
								s.c.Logger.Error(fmt.Sprintf("Error Joining Online Game: %s", err.Error()))
								continue
							} else {
								alreadyJoined = true
							}
						} else {
							// Avoid going in the same game again
							s.bot.TogglePause()
							// Reset the already joined check
							alreadyJoined = false
							continue
						}

					} else {
						s.c.Logger.Error(fmt.Sprintf("Error in Lobby Games loop. EnableLobbyGames: %t, CreateOnlineGames: %t, JoinOnlineGame: %t", s.c.CharacterCfg.Game.EnableLobbyGames, s.c.CharacterCfg.Game.CreateOnlineGames, s.c.CharacterCfg.Game.JoinOnlineGame))
						continue
					}
				} else {
					if err = s.c.Manager.NewGame(); err != nil {
						s.c.Logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
						continue
					}
				}
			}

			runs := s.runFactory.BuildRuns()
			gameStart := time.Now()
			if config.Characters[s.name].Game.RandomizeRuns {
				rand.Shuffle(len(runs), func(i, j int) { runs[i], runs[j] = runs[j], runs[i] })
			}
			event.Send(event.GameCreated(event.Text(s.name, "New game created"), "", ""))
			s.logGameStart(runs)
			err = s.bot.Run(ctx, firstRun, runs)
			if err != nil {
				if errors.Is(context.Canceled, ctx.Err()) {
					continue
				}

				switch {
				case errors.Is(err, health.ErrChicken):
					event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.c.Reader.Screenshot()), event.FinishedChicken))
					s.c.Logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
				case errors.Is(err, health.ErrMercChicken):
					event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.c.Reader.Screenshot()), event.FinishedMercChicken))
					s.c.Logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
				case errors.Is(err, health.ErrDied):
					event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.c.Reader.Screenshot()), event.FinishedDied))
					s.c.Logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
				default:
					event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.c.Reader.Screenshot()), event.FinishedError))
					s.c.Logger.Warn(
						fmt.Sprintf("Game finished with errors, reason: %s. Game total time: %0.2fs", err.Error(), time.Since(gameStart).Seconds()),
						slog.String("supervisor", s.name),
						slog.Uint64("mapSeed", uint64(s.c.Reader.CachedMapSeed)),
					)
				}
			}
			if exitErr := s.c.Manager.ExitGame(); exitErr != nil {
				errMsg := fmt.Sprintf("Error exiting game %s", err.Error())
				event.Send(event.GameFinished(event.WithScreenshot(s.name, errMsg, s.c.Reader.Screenshot()), event.FinishedError))

				return errors.New(errMsg)
			}
			firstRun = false
		}
	}
}
