package character

import (
	"errors"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"log/slog"
	"sort"
	"sync/atomic"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Berserker struct {
	BaseCharacter
	isKillingCouncil atomic.Bool
}

const (
	maxHorkRange      = 40
	meleeRange        = 5
	maxAttackAttempts = 20
)

var ErrNotInTravincal = errors.New("not in Travincal")

func (s *Berserker) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.BattleCommand, skill.BattleOrders, skill.Shout, skill.FindItem, skill.Berserk}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s *Berserker) IsKillingCouncil() bool {
	return s.isKillingCouncil.Load()
}

func (s *Berserker) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {

	for attackAttempts := 0; attackAttempts < maxAttackAttempts; attackAttempts++ {
		id, found := monsterSelector(*s.data)
		if !found {
			if !s.isKillingCouncil.Load() {
				s.FindItemOnNearbyCorpses(maxHorkRange)
			}
			return nil
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		monster, monsterFound := s.data.Monsters.FindByID(id)
		if !monsterFound || monster.Stats[stat.Life] <= 0 {
			continue
		}

		distance := s.pf.DistanceFromMe(monster.Position)
		if distance > meleeRange {
			err := step.MoveTo(monster.Position)
			if err != nil {
				s.logger.Warn("Failed to move to monster", slog.String("error", err.Error()))
				continue
			}
		}

		s.PerformBerserkAttack(monster.UnitID)
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

func (s *Berserker) PerformBerserkAttack(monsterID data.UnitID) {
	ctx := context.Get()
	monster, found := s.data.Monsters.FindByID(monsterID)
	if !found {
		return
	}

	// Ensure Berserk skill is active
	berserkKey, found := s.data.KeyBindings.KeyBindingForSkill(skill.Berserk)
	if found && s.data.PlayerUnit.RightSkill != skill.Berserk {
		ctx.HID.PressKeyBinding(berserkKey)
		time.Sleep(50 * time.Millisecond)
	}

	screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(monster.Position.X, monster.Position.Y)
	ctx.HID.Click(game.LeftButton, screenX, screenY)
}

func (s *Berserker) FindItemOnNearbyCorpses(maxRange int) {
	ctx := context.Get()

	findItemKey, found := s.data.KeyBindings.KeyBindingForSkill(skill.FindItem)
	if !found {
		s.logger.Debug("Find Item skill not found in key bindings")
		return
	}

	corpses := s.getSortedHorkableCorpses(s.data.Corpses, s.data.PlayerUnit.Position, maxRange)
	s.logger.Debug("Horkable corpses found", slog.Int("count", len(corpses)))

	for _, corpse := range corpses {
		err := step.MoveTo(corpse.Position)
		if err != nil {
			s.logger.Warn("Failed to move to corpse", slog.String("error", err.Error()))
			continue
		}

		if s.data.PlayerUnit.RightSkill != skill.FindItem {
			ctx.HID.PressKeyBinding(findItemKey)
			time.Sleep(time.Millisecond * 50)
		}

		clickPos := s.getOptimalClickPosition(corpse)
		screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(clickPos.X, clickPos.Y)
		ctx.HID.Click(game.RightButton, screenX, screenY)
		s.logger.Debug("Find Item used on corpse", slog.Any("corpse_id", corpse.UnitID))

		time.Sleep(time.Millisecond * 300)
	}

}

func (s *Berserker) getSortedHorkableCorpses(corpses data.Monsters, playerPos data.Position, maxRange int) []data.Monster {
	var horkableCorpses []data.Monster
	for _, corpse := range corpses {
		if s.isCorpseHorkable(corpse) && s.pf.DistanceFromMe(corpse.Position) <= maxRange {
			horkableCorpses = append(horkableCorpses, corpse)
		}
	}

	sort.Slice(horkableCorpses, func(i, j int) bool {
		distI := s.pf.DistanceFromMe(horkableCorpses[i].Position)
		distJ := s.pf.DistanceFromMe(horkableCorpses[j].Position)
		return distI < distJ
	})

	return horkableCorpses
}

func (s *Berserker) isCorpseHorkable(corpse data.Monster) bool {
	unhorkableStates := []state.State{
		state.CorpseNoselect,
		state.CorpseNodraw,
		state.Revive,
		state.Redeemed,
		state.Shatter,
		state.Freeze,
		state.Restinpeace,
	}

	for _, st := range unhorkableStates {
		if corpse.States.HasState(st) {
			return false
		}
	}

	return corpse.Type == data.MonsterTypeChampion ||
		corpse.Type == data.MonsterTypeMinion ||
		corpse.Type == data.MonsterTypeUnique ||
		corpse.Type == data.MonsterTypeSuperUnique
}

func (s *Berserker) getOptimalClickPosition(corpse data.Monster) data.Position {
	return data.Position{X: corpse.Position.X, Y: corpse.Position.Y + 1}
}

/*
todo need to define wich weapon slot I/II is primary/secondary before using that function
func (s *Berserker) swapToWeaponSlot(slot int) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "swapToWeaponSlot"

	currentSlot := "primary"
	if !s.isOnWeaponSlotOne() {
		currentSlot = "secondary"
	}

	ctx.Logger.Debug("Attempting weapon swap",
		slog.Int("targetSlot", slot),
		slog.String("currentSlot", currentSlot))

	if (slot == 1 && s.isOnWeaponSlotOne()) || (slot == 2 && !s.isOnWeaponSlotOne()) {
		ctx.Logger.Debug("Already on correct weapon slot", slog.Int("slot", slot))
		return nil
	}

	ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
	time.Sleep(100 * time.Millisecond)

	ctx.Logger.Debug("Weapon swap performed", slog.Int("toSlot", slot))
	return nil
}
func (s *Berserker) isOnWeaponSlotOne() bool {


}*/

func (s *Berserker) BuffSkills() []skill.ID {

	skillsList := make([]skill.ID, 0)
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.BattleCommand); found {
		skillsList = append(skillsList, skill.BattleCommand)
	}
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.Shout); found {
		skillsList = append(skillsList, skill.Shout)
	}
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.BattleOrders); found {
		skillsList = append(skillsList, skill.BattleOrders)
	}
	return skillsList
}

func (s *Berserker) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s *Berserker) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}
		return m.UnitID, true
	}, nil)
}

func (s *Berserker) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s *Berserker) KillAndariel() error {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s *Berserker) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s *Berserker) KillDuriel() error {
	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s *Berserker) KillMephisto() error {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s *Berserker) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.data.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
		if !found || diablo.Stats[stat.Life] <= 0 {
			if diabloFound {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		return s.killMonster(npc.Diablo, data.MonsterTypeNone)
	}
}

func (s *Berserker) KillCouncil() error {
	s.isKillingCouncil.Store(true)
	defer s.isKillingCouncil.Store(false)
	for {
		err := s.killAllCouncilMembers()
		if err != nil {
			if err == ErrNotInTravincal {
				s.logger.Info("Not in Travincal during Council kill, moving back")
				if moveErr := action.MoveToArea(area.Travincal); moveErr != nil {
					return moveErr
				}
				continue // Retry the Council kill after moving back to Travincal
			}
			return err
		}
		break // Exit the loop if killAllCouncilMembers completes without error
	}

	// Wait for corpses to settle
	time.Sleep(500 * time.Millisecond)

	// Perform horking in two passes
	for i := 0; i < 2; i++ {
		if err := s.ensureInTravincal(); err != nil {
			return err
		}

		s.FindItemOnNearbyCorpses(maxHorkRange)

		// Wait between passes
		time.Sleep(300 * time.Millisecond)

		// Refresh game data to catch any new corpses
		context.Get().RefreshGameData()
	}

	// Final wait for items to drop
	time.Sleep(500 * time.Millisecond)

	// Final item pickup
	err := action.ItemPickup(maxHorkRange)
	if err != nil {
		s.logger.Warn("Error during final item pickup after horking", "error", err)
		return err
	}

	// Wait a moment to ensure all items are picked up
	time.Sleep(300 * time.Millisecond)

	return nil
}

func (s *Berserker) killAllCouncilMembers() error {
	for {
		if err := s.ensureInTravincal(); err != nil {
			return ErrNotInTravincal
		}

		if !s.anyCouncilMemberAlive() {
			return nil
		}

		err := s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			for _, m := range d.Monsters.Enemies() {
				if (m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3) && m.Stats[stat.Life] > 0 {
					return m.UnitID, true
				}
			}
			return 0, false
		}, nil)

		if err != nil {
			// Check if we're still in Travincal after the attack sequence
			if checkErr := s.ensureInTravincal(); checkErr != nil {
				return ErrNotInTravincal
			}
			// If we're still in Travincal, return the original error
			return err
		}
	}
}

func (s *Berserker) anyCouncilMemberAlive() bool {
	for _, m := range s.data.Monsters.Enemies() {
		if (m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3) && m.Stats[stat.Life] > 0 {
			return true
		}
	}
	return false
}

// could be optimized it takes up to 15 second to get back in
func (s *Berserker) ensureInTravincal() error {
	ctx := context.Get()
	if ctx.Data.PlayerUnit.Area != area.Travincal {
		return ErrNotInTravincal
	}
	return nil
}

func (s *Berserker) KillIzual() error {
	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s *Berserker) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s *Berserker) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s *Berserker) KillBaal() error {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
