package action

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/itemfilter"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
)

func (b *Builder) Gamble() *Chain {
	return NewChain(func(d data.Data) (actions []Action) {
		if b.CharacterCfg.Gambling.Enabled && d.PlayerUnit.Stats[stat.StashGold] >= 2500000 {
			b.Logger.Info("Time to gamble! Visiting vendor...")

			openShopStep := step.KeySequence("home", "down", "down", "enter")
			vendorNPC := town.GetTownByArea(d.PlayerUnit.Area).GamblingNPC()

			// Jamella gamble button is the second one
			if vendorNPC == npc.Jamella {
				openShopStep = step.KeySequence("home", "down", "enter")
			}

			// Fix for Anya position
			if vendorNPC == npc.Drehya {
				actions = append(actions, b.MoveToCoords(data.Position{
					X: 5107,
					Y: 5119,
				}))
			}

			return append(actions,
				b.InteractNPC(vendorNPC,
					openShopStep,
					step.Wait(time.Second),
				),
				b.gambleItems(),
			)
		}

		return nil
	})
}

func (b *Builder) gambleItems() *StepChainAction {
	var itemBought data.Item
	currentIdx := 0
	lastStep := false
	return NewStepChain(func(d data.Data) []step.Step {
		if lastStep {
			if d.OpenMenus.Inventory {
				return []step.Step{step.SyncStep(func(d data.Data) error {
					b.HID.PressKey("esc")
					return nil
				})}
			}

			b.Logger.Info("Finished gambling", slog.Int("currentGold", d.PlayerUnit.TotalGold()))

			return nil
		}

		if itemBought.Name != "" {
			// For any reason item is still detected as located in vendor, I will take a look later into this
			for _, itm := range d.Items.ByLocation(item.LocationVendor) {
				if itm.UnitID == itemBought.UnitID {
					itemBought = itm
					itemBought.Location = item.LocationInventory
					b.Logger.Debug("Gambled for item", slog.Any("item", itemBought))
					break
				}
			}

			if itemfilter.Evaluate(itemBought, b.CharacterCfg.Runtime.Rules) {
				lastStep = true
				return []step.Step{step.Wait(time.Millisecond * 200)}
			} else {
				// Filter not pass, selling the item
				return []step.Step{step.SyncStep(func(d data.Data) error {
					b.sm.SellItem(itemBought)
					itemBought = data.Item{}
					return nil
				})}
			}
		}

		if d.PlayerUnit.TotalGold() < 500000 {
			lastStep = true
			return []step.Step{step.Wait(time.Millisecond * 200)}
		}

		for idx, itmName := range b.CharacterCfg.Gambling.Items {
			// Let's try to get one of each every time
			if currentIdx == len(b.CharacterCfg.Gambling.Items) {
				currentIdx = 0
			}

			if currentIdx > idx {
				continue
			}

			itm, found := d.Items.Find(itmName, item.LocationVendor)
			if !found {
				b.Logger.Debug("Item not found in gambling window, refreshing...", slog.String("item", string(itmName)))

				return []step.Step{step.SyncStep(func(d data.Data) error {
					b.HID.Click(game.LeftButton, ui.GambleRefreshButtonX, ui.GambleRefreshButtonY)
					return nil
				}),
					step.Wait(time.Millisecond * 500),
				}
			}

			return []step.Step{step.SyncStep(func(d data.Data) error {
				b.sm.BuyItem(itm, 1)
				itemBought = itm
				currentIdx++
				return nil
			})}
		}

		return nil
	}, RepeatUntilNoSteps())
}
