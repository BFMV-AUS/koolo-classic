package action

import (
	"errors"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
)

const (
	wpTabStartX     = 21.719
	wpTabStartY     = 7.928
	wpListStartX    = 5.37
	wpListStartY    = 6.852
	wpTabSize       = 69
	wpAreaBtnHeight = 49
)

func (b Builder) WayPoint(a area.Area) *StaticAction {
	allowedAreas := map[area.Area][2]int{
		area.StonyField:          {1, 3},
		area.BlackMarsh:          {1, 5},
		area.CatacombsLevel2:     {1, 9},
		area.LostCity:            {2, 6},
		area.ArcaneSanctuary:     {2, 8},
		area.LowerKurast:         {3, 5},
		area.Travincal:           {3, 8},
		area.DuranceOfHateLevel2: {3, 9},
		area.RiverOfFlame:        {4, 3},
		area.Harrogath:           {5, 1},
		area.FrigidHighlands:     {5, 2},
		area.HallsOfPain:         {5, 6},
	}

	return BuildStatic(func(data game.Data) (steps []step.Step) {
		// We don't need to move
		if data.PlayerUnit.Area == a {
			return
		}

		wpCoords, found := allowedAreas[a]
		if !found {
			panic("Area destination is not mapped on WayPoint Action (waypoint.go)")
		}

		for _, o := range data.Objects {
			if o.IsWaypoint() {
				steps = append(steps,
					step.InteractObject(o.Name, func(data game.Data) bool {
						return data.OpenMenus.Waypoint
					}),
					step.SyncStep(func(data game.Data) error {
						actTabX := int(float32(hid.GameAreaSizeX)/wpTabStartX) + (wpCoords[0]-1)*wpTabSize + (wpTabSize / 2)
						actTabY := int(float32(hid.GameAreaSizeY) / wpTabStartY)

						areaBtnX := int(float32(hid.GameAreaSizeX) / wpListStartX)
						areaBtnY := int(float32(hid.GameAreaSizeY)/wpListStartY) + (wpCoords[1]-1)*wpAreaBtnHeight + (wpAreaBtnHeight / 2)
						hid.MovePointer(actTabX, actTabY)
						helper.Sleep(200)
						hid.Click(hid.LeftButton)
						helper.Sleep(200)
						hid.MovePointer(areaBtnX, areaBtnY)
						helper.Sleep(200)
						hid.Click(hid.LeftButton)

						for i := 0; i < 10; i++ {
							d := b.gr.GetData(false)
							if d.PlayerUnit.Area == a {
								// Give some time to load the area
								helper.Sleep(4000)
								return nil
							}
							helper.Sleep(1000)
						}

						return errors.New("error changing area zone")
					}),
				)
			}
		}

		return
	}, Resettable())
}
