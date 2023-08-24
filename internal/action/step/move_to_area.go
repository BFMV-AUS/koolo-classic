package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type MoveToAreaStep struct {
	pathingStep
	area                  area.Area
	waitingForInteraction bool
	mouseOverAttempts     int
}

func MoveToLevel(area area.Area) *MoveToAreaStep {
	return &MoveToAreaStep{
		pathingStep: newPathingStep(),
		area:        area,
	}
}

func (m *MoveToAreaStep) Status(d data.Data) Status {
	if m.status == StatusCompleted {
		return StatusCompleted
	}

	// Give some extra time to render the UI
	if d.PlayerUnit.Area == m.area && time.Since(m.lastRun) > time.Second*1 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveToAreaStep) Run(d data.Data) error {
	if m.status == StatusNotStarted && helper.CanTeleport(d) {
		hid.PressKey(config.Config.Bindings.Teleport)
	}
	m.tryTransitionStatus(StatusInProgress)

	// Add some delay between clicks to let the character move to destination
	walkDuration := helper.RandomDurationMs(800, 2000)
	if !helper.CanTeleport(d) && time.Since(m.lastRun) < walkDuration {
		return nil
	}

	if helper.CanTeleport(d) && time.Since(m.lastRun) < config.Config.Runtime.CastDuration {
		return nil
	}

	if m.mouseOverAttempts > maxInteractions {
		return fmt.Errorf("area %d could not be interacted", m.area)
	}

	if (m.waitingForInteraction && time.Since(m.lastRun) < time.Second*1) || d.PlayerUnit.Area == m.area {
		return nil
	}

	m.lastRun = time.Now()
	for _, l := range d.AdjacentLevels {
		if l.Area == m.area {
			distance := pather.DistanceFromMe(d, l.Position)
			if distance > 5 {
				stuck := m.isPlayerStuck(d)
				if m.path == nil || !m.cachePath(d) || stuck {
					if stuck {
						tile := m.path.AstarPather[m.path.Distance()-1].(*pather.Tile)
						m.blacklistedPositions = append(m.blacklistedPositions, [2]int{tile.X, tile.Y})
					}
					path, _, found := pather.GetPath(d, l.Position)
					if !found {
						return errors.New("path could not be calculated, maybe there is an obstacle")
					}
					m.path = path
				}
				pather.MoveThroughPath(m.path, calculateMaxDistance(d, m.path, walkDuration), helper.CanTeleport(d))
				return nil
			}

			if l.IsEntrance {
				if d.HoverData.UnitType == 5 || d.HoverData.UnitType == 2 && d.HoverData.IsHovered {
					hid.Click(hid.LeftButton)
					m.waitingForInteraction = true
				}

				lx, ly := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, l.Position.X-2, l.Position.Y-2)
				x, y := helper.Spiral(m.mouseOverAttempts)
				hid.MovePointer(lx+x, ly+y)
				m.mouseOverAttempts++
				return nil
			} else {
				hid.PressKey(config.Config.Bindings.ForceMove)
			}
		}
	}

	return fmt.Errorf("area %s not found", m.area)
}
