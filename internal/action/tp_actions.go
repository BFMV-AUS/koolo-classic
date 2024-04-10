package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b *Builder) ReturnTown() *StepChainAction {
	return NewStepChain(func(d game.Data) (steps []step.Step) {
		if d.PlayerUnit.Area.IsTown() {
			return
		}

		return []step.Step{
			step.OpenPortal(d.CharacterCfg.Bindings.TP),
			step.InteractObject(object.TownPortal, func(d game.Data) bool {
				return d.PlayerUnit.Area.IsTown()
			}),
		}
	}, Resettable())
}

func (b *Builder) UsePortalInTown() *Chain {
	return NewChain(func(d game.Data) []Action {
		return []Action{b.UsePortalFrom(d.PlayerUnit.Name)}
	})
}

func (b *Builder) UsePortalFrom(owner string) *Chain {
	return NewChain(func(d game.Data) []Action {
		if !d.PlayerUnit.Area.IsTown() {
			return nil
		}

		for _, obj := range d.Objects {
			if obj.IsPortal() && obj.Owner == owner {
				tpArea := town.GetTownByArea(d.PlayerUnit.Area).TPWaitingArea(d)
				return []Action{
					b.MoveToCoords(tpArea),
					b.InteractObjectByID(obj.ID, func(d game.Data) bool {
						if !d.PlayerUnit.Area.IsTown() {
							helper.Sleep(500)
							return true
						}

						return false
					}),
				}
			}
		}

		return []Action{}
	})
}
