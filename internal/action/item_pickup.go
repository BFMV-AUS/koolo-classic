package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"strings"
)

func (b Builder) ItemPickup() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		itemsToPickup := b.getItemsToPickup(data)
		for _, item := range itemsToPickup {
			b.logger.Debug(fmt.Sprintf("Item Detected: %s [%s] at X:%d Y:%d", item.Name, item.Quality, item.Position.X, item.Position.Y))
			steps = append(steps, step.PickupItem(b.logger, item))
		}

		return
	}, CanBeSkipped())
}

func (b Builder) getItemsToPickup(data game.Data) []game.Item {
	missingHealingPotions := b.bm.GetMissingCount(data, game.HealingPotion)
	missingManaPotions := b.bm.GetMissingCount(data, game.ManaPotion)
	missingRejuvenationPotions := b.bm.GetMissingCount(data, game.RejuvenationPotion)
	var itemsToPickup []game.Item
	for _, item := range data.Items.Ground {
		for _, pickitItem := range config.Pickit.Items {
			if strings.EqualFold(item.Name, pickitItem.Name) {
				// Pickup potions only if they are required
				if item.IsHealingPotion() {
					if missingHealingPotions == 0 {
						break
					}
					itemsToPickup = append(itemsToPickup, item)
					missingHealingPotions--
					break
				}
				if item.IsManaPotion() {
					if missingManaPotions == 0 {
						break
					}
					itemsToPickup = append(itemsToPickup, item)
					missingManaPotions--
					break
				}
				if item.IsRejuvPotion() {
					if missingRejuvenationPotions == 0 {
						break
					}
					itemsToPickup = append(itemsToPickup, item)
					missingRejuvenationPotions--
					break
				}

				if b.shouldBePickedUp(item, pickitItem) {
					itemsToPickup = append(itemsToPickup, item)
					break
				}
			}

			// Check if we should pickup gold, based on amount
			if config.Pickit.PickupGold && strings.EqualFold(item.Name, "Gold") {
				if item.Stats[game.StatGold] >= config.Pickit.MinimumGoldToPickup {
					itemsToPickup = append(itemsToPickup, item)
					break
				}
			}
		}
	}

	return itemsToPickup
}

func (b Builder) shouldBePickedUp(i game.Item, pi config.ItemPickit) bool {
	if pi.Quality != "" && !strings.EqualFold(string(i.Quality), pi.Quality) {
		return false
	}

	if pi.Ethereal != nil && i.Ethereal != *pi.Ethereal {
		return false
	}

	pickup := true
	for stat, value := range i.Stats {
		for pickitStat, pickitValue := range pi.Stats {
			if strings.EqualFold(string(stat), pickitStat) {
				if value < pickitValue {
					pickup = false
					break
				}
			}
		}
	}

	return pickup
}
