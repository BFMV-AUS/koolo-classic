package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

var andarielStartingPosition = data.Position{
	X: 22561,
	Y: 9553,
}

type Andariel struct {
	baseRun
}

func (a Andariel) Name() string {
	return "Andariel"
}

func (a Andariel) BuildActions() (actions []action.Action) {
	// Moving to starting point (Catacombs Level 2)
	actions = append(actions, a.builder.WayPoint(area.CatacombsLevel2))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.CatacombsLevel3),
			step.MoveToLevel(area.CatacombsLevel4),
			step.MoveTo(andarielStartingPosition),
		}
	}))

	// Kill Andariel
	actions = append(actions, a.char.KillAndariel())
	return
}
