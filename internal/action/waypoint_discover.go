package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"log/slog"
)

func (b *Builder) DiscoverWaypoint() *Chain {
	return NewChain(func(d data.Data) []Action {
		b.Logger.Info("Trying to autodiscover Waypoint for current area", slog.Any("area", d.PlayerUnit.Area))
		for _, o := range d.Objects {
			if o.IsWaypoint() {
				return []Action{b.InteractObject(o.Name,
					func(d data.Data) bool {
						return d.OpenMenus.Waypoint
					},
					step.SyncStep(func(d data.Data) error {
						b.Logger.Info("Waypoint discovered", slog.Any("area", d.PlayerUnit.Area))
						helper.Sleep(500)
						b.HID.PressKey("esc")
						return nil
					}),
				)}
			}
		}

		b.Logger.Info("Waypoint not found :(", slog.Any("area", d.PlayerUnit.Area))
		return nil
	})
}
