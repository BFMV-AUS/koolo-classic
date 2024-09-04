package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	action2 "github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type Endugu struct {
	ctx *context.Status
}

func NewEndugu() *Endugu {
	return &Endugu{
		ctx: context.Get(),
	}
}

func (e Endugu) Name() string {
	return string(config.EnduguRun)
}

func (e Endugu) Run() error {

	// Use waypoint to FlayerJungle
	err := action2.WayPoint(area.FlayerJungle)
	if err != nil {
		return err
	}

	// Move to FlayerDungeonLevel1
	if err = action2.MoveToArea(area.FlayerDungeonLevel1); err != nil {
		return err
	}

	// Move to FlayerDungeonLevel2
	if err = action2.MoveToArea(area.FlayerDungeonLevel2); err != nil {
		return err
	}

	// Move to FlayerDungeonLevel3
	if err = action2.MoveToArea(area.FlayerDungeonLevel3); err != nil {
		return err
	}

	var khalimChest2 data.Object

	// Move to KhalimChest
	action2.MoveTo(func() (data.Position, bool) {
		for _, o := range e.ctx.Data.Objects {
			if o.Name == object.KhalimChest2 {
				khalimChest2 = o
				return o.Position, true
			}
		}
		return data.Position{}, false
	})

	// Clear monsters around player
	action2.ClearAreaAroundPlayer(15, data.MonsterEliteFilter())

	// Open the chest
	return action2.InteractObject(khalimChest2, func() bool {
		for _, obj := range e.ctx.Data.Objects {
			if obj.Name == object.KhalimChest2 && !obj.Selectable {
				return true
			}
		}
		return false
	})
}