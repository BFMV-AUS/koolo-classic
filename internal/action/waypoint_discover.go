package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

func (b Builder) DiscoverWaypoint() *Factory {
	interacted := false

	return NewFactory(func(d data.Data) Action {
		b.logger.Info("Trying to autodiscover Waypoint for current area", zap.Any("area", d.PlayerUnit.Area))
		if interacted {
			return nil
		}

		for _, o := range d.Objects {
			if o.IsWaypoint() {
				pos := o.Position
				if d.PlayerUnit.Area == area.RiverOfFlame {
					pos = data.Position{X: 7800, Y: 5919}
				}

				if pather.DistanceFromMe(d, pos) < 15 {
					return BuildStatic(func(d data.Data) []step.Step {
						return []step.Step{
							step.MoveTo(pos),
							step.InteractObject(o.Name, func(d data.Data) bool {
								return d.OpenMenus.Waypoint
							}),
							step.SyncStep(func(d data.Data) error {
								helper.Sleep(1000)
								hid.PressKey("esc")
								interacted = true
								return nil
							}),
						}
					})
				}

				return b.MoveTo(func(d data.Data) (data.Position, bool) {
					for _, o := range d.Objects {
						if o.IsWaypoint() {
							return pos, true
						}
					}

					return data.Position{}, false
				})
			}
		}

		b.logger.Info("Waypoint not found :(", zap.Any("area", d.PlayerUnit.Area))
		return nil
	})
}
