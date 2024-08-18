package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/v2/action"
	"github.com/hectorgimenez/koolo/internal/v2/context"
)

type SpiderCavern struct {
	ctx *context.Status
}

func NewSpiderCavern() *Tristram {
	return &Tristram{
		ctx: context.Get(),
	}
}

func (run SpiderCavern) Name() string {
	return string(config.TristramRun)
}

func (run SpiderCavern) Run() error {

	// Define a default monster filter
	monsterFilter := data.MonsterAnyFilter()

	// Update filter if we selected to clear only elites
	if run.ctx.CharacterCfg.Game.DrifterCavern.FocusOnElitePacks {
		monsterFilter = data.MonsterEliteFilter()
	}

	// Use waypoint to Spider Forest
	err := action.WayPoint(area.SpiderForest)
	if err != nil {
		return err
	}

	// Move to the correct area
	if err = action.MoveToArea(area.SpiderCavern); err != nil {
		return err
	}

	// Clear the area
	return action.ClearCurrentLevel(run.ctx.CharacterCfg.Game.SpiderCavern.OpenChests, monsterFilter)
}
