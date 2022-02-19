package game

const (
	// Towns
	AreaRogueEncampment Area = "RogueEncampment"
	AreaLutGholein      Area = "LutGholein"
	AreaKurastDocks     Area = "KurastDocks"
	AreaPandemonium     Area = "ThePandemoniumFortress"
	AreaHarrogath       Area = "Harrogath"

	AreaCatacombsLevel2     Area = "CatacombsLevel2"
	AreaCatacombsLevel3     Area = "CatacombsLevel3"
	AreaCatacombsLevel4     Area = "CatacombsLevel4"
	AreaNihlathaksTemple    Area = "NihlathaksTemple"
	AreaDuranceOfHateLevel2 Area = "DuranceOfHateLevel2"
	AreaDuranceOfHateLevel3 Area = "DuranceOfHateLevel3"
	AreaBlackMarsh          Area = "BlackMarsh"
	AreaForgottenTower      Area = "ForgottenTower"
	AreaTowerCellarLevel1   Area = "TowerCellarLevel1"
	AreaTowerCellarLevel2   Area = "TowerCellarLevel2"
	AreaTowerCellarLevel3   Area = "TowerCellarLevel3"
	AreaTowerCellarLevel4   Area = "TowerCellarLevel4"
	AreaTowerCellarLevel5   Area = "TowerCellarLevel5"
	ArcaneSanctuary         Area = "ArcaneSanctuary"

	// Classes
	ClassSorceress Class = "Sorceress"

	// Skills
	SkillBattleOrders Skill = "BattleOrders"
)

type Data struct {
	Health         Health
	Area           Area
	AreaOrigin     Position
	Corpse         Corpse
	Monsters       map[NPCID]Monster
	CollisionGrid  [][]int
	PlayerUnit     PlayerUnit
	NPCs           map[NPCID]NPC
	Items          Items
	Objects        []Object
	AdjacentLevels []Level
	OpenMenus      OpenMenus
}

type Area string

type Level struct {
	Area     Area
	Position Position
}

func (a Area) IsTown() bool {
	switch a {
	case AreaRogueEncampment, AreaLutGholein, AreaKurastDocks, AreaPandemonium, AreaHarrogath:
		return true
	}

	return false
}

type Class string
type Corpse struct {
	Found     bool
	IsHovered bool
	Position  Position
}
type Monster struct {
	Name      string
	IsHovered bool
	Position  Position
}

type Position struct {
	X int
	Y int
}

type Skill string
type PlayerUnit struct {
	Name      string
	IsHovered bool
	Position  Position
	Stats     map[Stat]int
	Skills    map[Skill]int
	Class     Class
}

type NPC struct {
	Name      string
	Positions []Position
}

type OpenMenus struct {
	Inventory   bool
	NPCInteract bool
	NPCShop     bool
	Stash       bool
	Waypoint    bool
}
