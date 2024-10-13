package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	ho_sorceressMaxAttacksLoop = 40
	ho_sorceressMinDistance    = 0
	ho_sorceressMaxDistance    = 30
)

type HydraOrbSorceress struct {
	BaseCharacter
}

func (s HydraOrbSorceress) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.FrozenOrb, skill.Hydra, skill.Teleport, skill.TomeOfTownPortal, skill.ShiverArmor, skill.StaticField}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(cskill); !found {
			switch cskill {
			// Since we can have one of 3 armors:
			case skill.ShiverArmor:
				_, found1 := s.data.KeyBindings.KeyBindingForSkill(skill.FrozenArmor)
				_, found2 := s.data.KeyBindings.KeyBindingForSkill(skill.ChillingArmor)
				if !found1 && !found2 {
					missingKeybindings = append(missingKeybindings, skill.ShiverArmor)
				}
			default:
				missingKeybindings = append(missingKeybindings, cskill)
			}
		}
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s HydraOrbSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	// previousSelfHydra := time.Time{}

	for {
		id, found := monsterSelector(*s.data)
		if !found {
			return nil
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return nil
		}

		monster, found := s.data.Monsters.FindByID(id)
		if !found {
			s.logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		opts := step.Distance(ho_sorceressMinDistance, ho_sorceressMaxDistance)

		// Cast a Blizzard on very close mobs, in order to clear possible trash close the player, every two attack rotations
		// if time.Since(previousSelfHydra) > time.Second*4 && !s.data.PlayerUnit.States.HasState(state.Cooldown) {
		//	for _, m := range s.data.Monsters.Enemies() {
		//		if dist := s.pf.DistanceFromMe(m.Position); dist < 4 {
		//			s.logger.Debug("Monster detected close to the player, casting Hydra on myself")
		//			previousSelfHydra = time.Now()
		//			step.SecondaryAttack(skill.Hydra, m.UnitID, 1, opts)
		//		}
		//	}
		//}

		if s.data.PlayerUnit.States.HasState(state.Cooldown) {
			step.SecondaryAttack(skill.Hydra, id, 1, opts)
		}

		step.SecondaryAttack(skill.FrozenOrb, id, 1, opts)

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s HydraOrbSorceress) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s HydraOrbSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, _ int, _ bool, skipOnImmunities []stat.Resist) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s HydraOrbSorceress) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	armors := []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor}
	for _, armor := range armors {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			return skillsList
		}
	}

	return skillsList
}

func (s HydraOrbSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s HydraOrbSorceress) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, ho_sorceressMaxDistance, false, nil)
}

func (s HydraOrbSorceress) KillAndariel() error {
	for {
		boss, found := s.data.Monsters.FindOne(npc.Andariel, data.MonsterTypeUnique)
		if !found || boss.Stats[stat.Life] <= 0 {
			return nil // Andariel is dead or not found
		}

		err := s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, ho_sorceressMaxDistance, false, nil)
		if err != nil {
			return err
		}

		// Short delay before checking again
		time.Sleep(100 * time.Millisecond)
	}
}
func (s HydraOrbSorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, ho_sorceressMaxDistance, false, nil)
}

func (s HydraOrbSorceress) KillDuriel() error {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, ho_sorceressMaxDistance, true, nil)
}

func (s HydraOrbSorceress) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		var veryImmunes []data.Monster
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				if m.IsImmune(stat.ColdImmune) && m.IsImmune(stat.FireImmune) {
					veryImmunes = append(veryImmunes, m)
				} else {
					councilMembers = append(councilMembers, m)
				}
			}
		}

		councilMembers = append(councilMembers, veryImmunes...)

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s HydraOrbSorceress) KillMephisto() error {
	for {
		boss, found := s.data.Monsters.FindOne(npc.Mephisto, data.MonsterTypeUnique)
		if !found || boss.Stats[stat.Life] <= 0 {
			return nil // Mephisto is dead or not found
		}

		err := s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, sorceressMaxDistance, true, nil)
		if err != nil {
			return err
		}

		// Short delay before checking again
		time.Sleep(100 * time.Millisecond)
	}
}

func (s HydraOrbSorceress) KillIzual() error {
	m, _ := s.data.Monsters.FindOne(npc.Izual, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(5, 8))

	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
}

func (s HydraOrbSorceress) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			time.Sleep(200)
			continue
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		_ = step.SecondaryAttack(skill.StaticField, diablo.UnitID, 5, step.Distance(3, 8))

		return s.killMonster(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s HydraOrbSorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, ho_sorceressMaxDistance, false, s.cfg.Game.Pindleskin.SkipOnImmunities)
}

func (s HydraOrbSorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, ho_sorceressMaxDistance, false, nil)
}

func (s HydraOrbSorceress) KillBaal() error {
	m, _ := s.data.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeUnique)
	step.SecondaryAttack(skill.StaticField, m.UnitID, 5, step.Distance(5, 8))

	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
