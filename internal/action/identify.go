package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/stats"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) IdentifyAll() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		b.logger.Info("Identifying items...")
		steps = append(steps,
			step.SyncStep(func(data game.Data) error {
				hid.PressKey(config.Config.Bindings.OpenInventory)
				b.identifyItems(data)
				hid.PressKey("esc")

				return nil
			}),
		)

		return
	}, CanBeSkipped())
}

func (b Builder) identifyItems(data game.Data) {
	// This workaround is to force to not identify items if it's the first game, to prevent identifying not wanted items
	for _, s := range stats.Status.RunStats {
		if s.TotalRunsTime == 0 {
			return
		}
	}

	idTome, found := getIDTome(data)
	if !found {
		b.logger.Warn("ID Tome not found, not identifying items")
		return
	}

	xIDTome := int(float32(hid.GameAreaSizeX)/town.InventoryTopLeftX) + idTome.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
	yIDTome := int(float32(hid.GameAreaSizeY)/town.InventoryTopLeftY) + idTome.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
	for _, i := range data.Items.Inventory {
		if i.Identified || i.Quality == game.ItemQualityNormal || i.Quality == game.ItemQualitySuperior {
			continue
		}

		hid.MovePointer(xIDTome, yIDTome)
		helper.Sleep(100)
		hid.Click(hid.RightButton)
		x := int(float32(hid.GameAreaSizeX)/town.InventoryTopLeftX) + i.Position.X*town.ItemBoxSize + (town.ItemBoxSize / 2)
		y := int(float32(hid.GameAreaSizeY)/town.InventoryTopLeftY) + i.Position.Y*town.ItemBoxSize + (town.ItemBoxSize / 2)
		hid.MovePointer(x, y)
		helper.Sleep(100)
		hid.Click(hid.LeftButton)
		helper.Sleep(350)
	}
}

func getIDTome(data game.Data) (game.Item, bool) {
	for _, i := range data.Items.Inventory {
		if i.Name == game.ItemTomeOfIdentify {
			return i, true
		}
	}

	return game.Item{}, false
}
