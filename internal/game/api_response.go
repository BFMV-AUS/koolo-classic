package game

type statusHttpResponse struct {
	Life    int `json:"life"`
	MaxLife int `json:"max_life"`
	Mana    int `json:"mana"`
	MaxMana int `json:"max_mana"`
	Merc    struct {
		Alive   bool `json:"alive"`
		Life    int  `json:"life"`
		MaxLife int  `json:"max_life"`
	} `json:"merc"`
}

type requestSucceed struct {
	Success bool `json:"success"`
}

type gameDataHttpResponse struct {
	Success    bool               `json:"success"`
	Area       string             `json:"area"`
	AreaOrigin position           `json:"area_origin"`
	Status     statusHttpResponse `json:"status"`
	Corpses    []struct {
		Name      string   `json:"name"`
		IsHovered bool     `json:"is_hovered"`
		Position  position `json:"position"`
	} `json:"corpses"`
	PlayerUnit struct {
		Name     string   `json:"name"`
		Position position `json:"position"`
		Skills   []struct {
			Skill  string `json:"skill"`
			Points int    `json:"points"`
		} `json:"skills"`
		Stats []struct {
			Stat  string `json:"stat"`
			Value int    `json:"value"`
		} `json:"stats"`
		PlayerClass string `json:"player_class"`
	} `json:"player_unit"`
	Items []struct {
		ID        int      `json:"id"`
		Position  position `json:"position"`
		Name      string   `json:"name"`
		IsHovered bool     `json:"is_hovered"`
		Place     string   `json:"place"`
		Quality   string   `json:"quality"`
		Ethereal  bool     `json:"ethereal"`
		Stats     []struct {
			Stat  string `json:"stat"`
			Value int    `json:"value"`
		} `json:"stats"`
	} `json:"items"`
	Objects []struct {
		Position   position `json:"position"`
		Name       string   `json:"name"`
		IsHovered  bool     `json:"is_hovered"`
		Selectable bool     `json:"selectable"`
	} `json:"objects"`
	Monsters []struct {
		Position  position `json:"position"`
		Name      string   `json:"name"`
		IsHovered bool     `json:"is_hovered"`
	} `json:"monsters"`
	NPCs []struct {
		Name      string     `json:"name"`
		Positions []position `json:"positions"`
	}
	CollisionGrid [][]int `json:"collision_grid"`
	MenuOpen      struct {
		Inventory   bool `json:"inventory"`
		NPCInteract bool `json:"npc_interact"`
		NPCShop     bool `json:"npc_shop"`
		Stash       bool `json:"stash"`
		Waypoint    bool `json:"waypoint"`
	} `json:"menu_open"`
	AdjacentLevels []struct {
		Area      string     `json:"area"`
		Positions []position `json:"positions"`
		IsPortal  bool
	} `json:"adjacent_levels"`
}

type position struct {
	X float32 `json:"X"`
	Y float32 `json:"Y"`
}
