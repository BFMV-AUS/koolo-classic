package character

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	sorceressMaxAttacksLoop = 10
	sorceressMinDistance    = 25
	sorceressMaxDistance    = 30
)

type BlizzardSorceress struct {
	BaseCharacter
}

func (s BlizzardSorceress) KillMonsterSequence(
	monsterSelector func(d data.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	completedAttackLoops := 0
	previousUnitID := 0

	return action.NewStepChain(func(d data.Data) []step.Step {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		if len(opts) == 0 {
			opts = append(opts, step.Distance(sorceressMinDistance, sorceressMaxDistance))
		}
		//if useStaticField {
		//	steps = append(steps,
		//		step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, time.Millisecond*100, step.Distance(sorceressMinDistance, maxDistance)),
		//		step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, id, 5, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, 15)),
		//	)
		//}
		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)
		// Cast a Blizzard on very close mobs, in order to clear possible trash close the player, every two attack rotations
		if completedAttackLoops%2 == 0 {
			for _, m := range d.Monsters.Enemies() {
				if d := pather.DistanceFromMe(d, m.Position); d < 4 {
					s.logger.Debug("Monster detected close to the player, casting Blizzard over it")
					steps = append(steps, step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, m.UnitID, 1, opts...))
					break
				}
			}
		}

		// In case monster is stuck behind a wall or character is not able to reachh it we will short the distance
		if completedAttackLoops > 5 {
			s.logger.Debug("Looks like monster is not reachable, moving closer")
			opts = append(opts, step.Distance(2, 8))
		}

		steps = append(steps,
			step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, opts...),
			step.PrimaryAttack(id, 4, opts...),
		)
		completedAttackLoops++
		previousUnitID = int(id)

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s BlizzardSorceress) Buff() action.Action {
	return action.NewStepChain(func(d data.Data) (steps []step.Step) {
		steps = append(steps, s.buffCTA()...)
		steps = append(steps, step.SyncStep(func(d data.Data) error {
			if config.Config.Bindings.Sorceress.FrozenArmor != "" {
				hid.PressKey(config.Config.Bindings.Sorceress.FrozenArmor)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
			}

			return nil
		}))

		return
	})
}

func (s BlizzardSorceress) KillCountess() action.Action {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillAndariel() action.Action {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeNone, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillSummoner() action.Action {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeNone, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillDuriel() action.Action {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeNone, sorceressMaxDistance, true, nil)
}

func (s BlizzardSorceress) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, skipOnImmunities)
}

func (s BlizzardSorceress) KillMephisto() action.Action {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeNone, sorceressMaxDistance, true, nil)
}

func (s BlizzardSorceress) KillNihlathak() action.Action {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillDiablo() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				return []step.Step{step.Wait(time.Second * 20)}
			}),
			action.NewChain(func(d data.Data) []action.Action {
				diablo, found := d.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
				if !found {
					return nil
				}

				s.logger.Info("Diablo detected, attacking")
				return []action.Action{
					action.NewStepChain(func(d data.Data) []step.Step {

						return []step.Step{
							step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, diablo.UnitID, 8, step.Distance(1, 5)),
						}
					}),
					s.killMonster(npc.Diablo, data.MonsterTypeNone),
				}
			}),
		}
	})
}

func (s BlizzardSorceress) KillIzual() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				mephisto, _ := d.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, mephisto.UnitID, 7, step.Distance(1, 4)),
				}
			}),
			// We will need a lot of cycles to kill him probably
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
		}
	})
}

func (s BlizzardSorceress) KillBaal() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				mephisto, _ := d.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, mephisto.UnitID, 7, step.Distance(1, 4)),
				}
			}),
			// We will need a lot of cycles to kill him probably
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
		}
	})
}

func (s BlizzardSorceress) KillCouncil() action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		var coldImmunes []data.Monster
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				if m.IsImmune(stat.ColdImmune) {
					coldImmunes = append(coldImmunes, m)
				} else {
					councilMembers = append(councilMembers, m)
				}
			}
		}

		councilMembers = append(councilMembers, coldImmunes...)

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil, step.Distance(8, sorceressMaxDistance))
}

func (s BlizzardSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, maxDistance int, useStaticField bool, skipOnImmunities []stat.Resist) action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities, step.Distance(sorceressMinDistance, maxDistance))
}

func (s BlizzardSorceress) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}
