package run

import (
	"context"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/town"
	"strings"
	"time"
)

type Companion struct {
	baseRun
}

func (s Companion) Name() string {
	return "Companion"
}

func (s Companion) BuildActions() []action.Action {
	var lastInteractionEvent *event.InteractedToEvent
	var leaderUnitIDTarget data.UnitID
	tpRequested := false
	waitingForLeaderSince := time.Time{}

	// TODO: Deregister this listener or will leak
	s.EventListener.Register(func(ctx context.Context, e event.Event) error {
		if config.Characters[e.Supervisor()].CharacterName == s.CharacterCfg.Companion.LeaderName {
			if evt, ok := e.(event.CompanionLeaderAttackEvent); ok {
				leaderUnitIDTarget = evt.TargetUnitID
			}

			if evt, ok := e.(event.InteractedToEvent); ok {
				lastInteractionEvent = &evt
			}
		}

		return nil
	})

	return []action.Action{
		action.NewChain(func(d data.Data) []action.Action {
			leaderRosterMember, _ := d.Roster.FindByName(s.CharacterCfg.Companion.LeaderName)

			// Leader is NOT in the same act, so we will try to change to the corresponding act
			if leaderRosterMember.Area.Act() != d.PlayerUnit.Area.Act() {
				// Follower is NOT in town
				if !d.PlayerUnit.Area.IsTown() {

					// Portal is found nearby
					if _, foundPortal := getClosestPortal(d, leaderRosterMember.Name); foundPortal {
						return []action.Action{
							s.builder.UsePortalFrom(leaderRosterMember.Name),
						}
					}

					// Portal is not found nearby
					if hasEnoughPortals(d) {
						return []action.Action{
							s.builder.ReturnTown(),
						}
					}

					// there is NO portal open and follower does NOT have enough portals. Just exit
					return []action.Action{}
				}

				// Follower is in town. Just change the act
				return []action.Action{
					s.builder.WayPoint(town.GetTownByArea(leaderRosterMember.Area).TownArea()),
				}
			}

			if lastInteractionEvent != nil {
				switch lastInteractionEvent.InteractionType {
				case event.InteractionTypeEntrance:
					a := area.Area(lastInteractionEvent.ID)
					lastInteractionEvent = nil
					if !d.PlayerUnit.Area.IsTown() {
						return []action.Action{
							s.builder.MoveToArea(a),
						}
					}
				case event.InteractionTypeObject:
					oName := object.Name(lastInteractionEvent.ID)
					lastInteractionEvent = nil
					o, found := d.Objects.FindOne(oName)
					if found && ((o.IsWaypoint() && !d.PlayerUnit.Area.IsTown()) || o.IsRedPortal()) {
						return []action.Action{
							s.builder.InteractObject(oName, func(dat data.Data) bool {
								if o.IsWaypoint() {
									return dat.OpenMenus.Waypoint
								}

								return d.PlayerUnit.Area != dat.PlayerUnit.Area
							}),
						}
					}
				case event.InteractionTypeNPC:
					npcID := npc.ID(lastInteractionEvent.ID)
					lastInteractionEvent = nil
					switch npcID {
					case npc.Warriv, npc.Meshif:
						return []action.Action{
							s.builder.ReturnTown(),
							s.builder.InteractNPC(npcID, step.KeySequence("home", "down", "enter")),
						}
					}
				}
			}

			// Is leader too far away?
			if pather.DistanceFromMe(d, leaderRosterMember.Position) > 100 {
				// In some cases this "follower in town -> use portal -> follower outside town -> use portal"
				// loop can go on forever. But it is responsibility of a leader to not cause it...

				// Follower in town
				if d.PlayerUnit.Area.IsTown() {
					// Request a TP
					if pather.DistanceFromMe(d, town.GetTownByArea(d.PlayerUnit.Area).TPWaitingArea(d)) < 10 && !tpRequested {
						event.Send(event.CompanionRequestedTP(event.Text(s.Supervisor, "TP Requested")))
						tpRequested = true
						return []action.Action{
							s.builder.Wait(time.Second),
						}
					}

					if _, foundPortal := getClosestPortal(d, leaderRosterMember.Name); foundPortal && !leaderRosterMember.Area.IsTown() {
						tpRequested = false
						return []action.Action{
							s.builder.UsePortalFrom(leaderRosterMember.Name),
						}
					}

					// Go to TP waiting area
					return []action.Action{
						s.builder.MoveToCoords(town.GetTownByArea(d.PlayerUnit.Area).TPWaitingArea(d)),
					}
				}

				// Otherwise just wait
				if waitingForLeaderSince.IsZero() || time.Since(waitingForLeaderSince) > time.Second*5 {
					waitingForLeaderSince = time.Now()
				}

				if time.Since(waitingForLeaderSince) > time.Second*3 {
					return []action.Action{
						s.builder.ReturnTown(),
					}
				}

				return []action.Action{
					s.builder.Wait(100),
				}
			}

			waitingForLeaderSince = time.Time{}
			// If distance from leader is acceptable and is attacking, support him
			distanceFromMe := pather.DistanceFromMe(d, leaderRosterMember.Position)
			if distanceFromMe < 30 {
				monster, found := d.Monsters.FindByID(leaderUnitIDTarget)
				if s.CharacterCfg.Companion.Attack && found {
					return []action.Action{s.killMonsterInCompanionMode(monster)}
				}

				// If there is no monster to attack, and we are close enough to the leader just wait
				if distanceFromMe < 4 {
					return []action.Action{
						s.builder.ItemPickup(false, 8),
						s.builder.Wait(100),
					}
				}
			}

			return []action.Action{
				action.NewStepChain(func(d data.Data) []step.Step {
					return []step.Step{step.MoveTo(s.CharacterCfg, leaderRosterMember.Position, step.WithTimeout(time.Millisecond*500))}
				}),
			}
		}, action.RepeatUntilNoSteps()),
	}
}

func getClosestPortal(d data.Data, leaderName string) (*data.Object, bool) {
	for _, o := range d.Objects {
		if o.IsPortal() && pather.DistanceFromMe(d, o.Position) <= 40 && strings.EqualFold(o.Owner, leaderName) {
			return &o, true
		}
	}

	return nil, false
}

func hasEnoughPortals(d data.Data) bool {
	portalTome, pFound := d.Items.Find(item.TomeOfTownPortal, item.LocationInventory)
	if pFound {
		return portalTome.Stats[stat.Quantity].Value > 0
	}

	return false
}

func (s Companion) killMonsterInCompanionMode(m data.Monster) action.Action {
	switch m.Name {
	case npc.Andariel:
		return s.char.KillAndariel()
	case npc.Duriel:
		return s.char.KillDuriel()
	case npc.Mephisto:
		return s.char.KillMephisto()
	case npc.Diablo:
		return s.char.KillDiablo()
	}

	return s.char.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		return m.UnitID, true
	}, nil)
}
