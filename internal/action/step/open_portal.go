package step

import (
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"

	"github.com/hectorgimenez/koolo/internal/helper"
)

type OpenPortalStep struct {
	basicStep
}

func OpenPortal() *OpenPortalStep {
	return &OpenPortalStep{
		basicStep: newBasicStep(),
	}
}

func (s *OpenPortalStep) Status(d game.Data, _ container.Container) Status {
	if s.status == StatusCompleted {
		return StatusCompleted
	}

	// Give some extra time, sometimes if we move the mouse over the portal before is shown
	// and there is an intractable entity behind it, will keep it focused
	if time.Since(s.LastRun()) > time.Second*1 {
		for _, o := range d.Objects {
			if o.IsPortal() {
				return s.tryTransitionStatus(StatusCompleted)
			}
		}
	}

	return StatusInProgress
}

func (s *OpenPortalStep) Run(d game.Data, container container.Container) error {
	// Give some time to portal to popup before retrying...
	if time.Since(s.LastRun()) < time.Second*2 {
		return nil
	}

	container.HID.PressKeyBinding(d.KeyBindings.MustKBForSkill(skill.TomeOfTownPortal))
	helper.Sleep(250)
	container.HID.Click(game.RightButton, 300, 300)
	s.lastRun = time.Now()

	return nil
}
