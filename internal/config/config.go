package config

import (
	"errors"
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	cp "github.com/otiai10/copy"
	"os"
	"strings"
	"time"

	"github.com/hectorgimenez/d2go/pkg/nip"

	"gopkg.in/yaml.v3"
)

var (
	Koolo      *KooloCfg
	Characters map[string]*CharacterCfg
)

type KooloCfg struct {
	FirstRun          bool `yaml:"firstRun"`
	CheckGameSettings bool `yaml:"checkGameSettings"`
	Debug             struct {
		Log       bool `yaml:"log"`
		RenderMap bool `yaml:"renderMap"`
	} `yaml:"debug"`
	LogSaveDirectory string `yaml:"logSaveDirectory"`
	D2LoDPath        string `yaml:"D2LoDPath"`
	D2RPath          string `yaml:"D2RPath"`
	Discord          struct {
		Enabled   bool   `yaml:"enabled"`
		ChannelID string `yaml:"channelId"`
		Token     string `yaml:"token"`
	} `yaml:"discord"`
	Telegram struct {
		Enabled bool   `yaml:"enabled"`
		ChatID  int64  `yaml:"chatId"`
		Token   string `yaml:"token"`
	}
}

type CharacterCfg struct {
	MaxGameLength int    `yaml:"maxGameLength"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	Realm         string `yaml:"realm"`
	Health        struct {
		HealingPotionAt     int `yaml:"healingPotionAt"`
		ManaPotionAt        int `yaml:"manaPotionAt"`
		RejuvPotionAtLife   int `yaml:"rejuvPotionAtLife"`
		RejuvPotionAtMana   int `yaml:"rejuvPotionAtMana"`
		MercHealingPotionAt int `yaml:"mercHealingPotionAt"`
		MercRejuvPotionAt   int `yaml:"mercRejuvPotionAt"`
		ChickenAt           int `yaml:"chickenAt"`
		MercChickenAt       int `yaml:"mercChickenAt"`
	} `yaml:"health"`
	Bindings struct {
		OpenInventory       string `yaml:"openInventory"`
		OpenCharacterScreen string `yaml:"openCharacterScreen"`
		OpenSkillTree       string `yaml:"openSkillTree"`
		OpenQuestLog        string `yaml:"openQuestLog"`
		Potion1             string `yaml:"potion1"`
		Potion2             string `yaml:"potion2"`
		Potion3             string `yaml:"potion3"`
		Potion4             string `yaml:"potion4"`
		ForceMove           string `yaml:"forceMove"`
		StandStill          string `yaml:"standStill"`
		SwapWeapon          string `yaml:"swapWeapon"`
		Teleport            string `yaml:"teleport"`
		TP                  string `yaml:"tp"`
		CTABattleCommand    string `yaml:"CTABattleCommand"`
		CTABattleOrders     string `yaml:"CTABattleOrders"`

		// Class Specific bindings
		Sorceress struct {
			Blizzard     string `yaml:"blizzard"`
			StaticField  string `yaml:"staticField"`
			FrozenArmor  string `yaml:"frozenArmor"`
			FireBall     string `yaml:"fireBall"`
			Nova         string `yaml:"nova"`
			EnergyShield string `yaml:"energyShield"`
		} `yaml:"sorceress"`
		Paladin struct {
			Concentration string `yaml:"concentration"`
			HolyShield    string `yaml:"holyShield"`
			Vigor         string `yaml:"vigor"`
			Redemption    string `yaml:"redemption"`
			Cleansing     string `yaml:"cleansing"`
		} `yaml:"paladin"`
	} `yaml:"bindings"`
	Inventory struct {
		InventoryLock [][]int `yaml:"inventoryLock"`
		BeltColumns   struct {
			Healing      int `yaml:"healing"`
			Mana         int `yaml:"mana"`
			Rejuvenation int `yaml:"rejuvenation"`
		} `yaml:"beltColumns"`
	} `yaml:"inventory"`
	Character struct {
		Class         string `yaml:"class"`
		CastingFrames int    `yaml:"castingFrames"`
		UseMerc       bool   `yaml:"useMerc"`
	} `yaml:"character"`
	Game struct {
		ClearTPArea   bool                  `yaml:"clearTPArea"`
		Difficulty    difficulty.Difficulty `yaml:"difficulty"`
		RandomizeRuns bool                  `yaml:"randomizeRuns"`
		Runs          []string              `yaml:"runs"`
		Pindleskin    struct {
			SkipOnImmunities []stat.Resist `yaml:"skipOnImmunities"`
		} `yaml:"pindleskin"`
		Mephisto struct {
			KillCouncilMembers bool `yaml:"killCouncilMembers"`
			OpenChests         bool `yaml:"openChests"`
		} `yaml:"mephisto"`
		Tristram struct {
			ClearPortal       bool `yaml:"clearPortal"`
			FocusOnElitePacks bool `yaml:"focusOnElitePacks"`
		} `yaml:"tristram"`
		Nihlathak struct {
			ClearArea bool `yaml:"clearArea"`
		} `yaml:"nihlathak"`
		Baal struct {
			KillBaal bool `yaml:"killBaal"`
		} `yaml:"baal"`
		TerrorZone struct {
			FocusOnElitePacks bool          `yaml:"focusOnElitePacks"`
			SkipOnImmunities  []stat.Resist `yaml:"skipOnImmunities"`
			SkipOtherRuns     bool          `yaml:"skipOtherRuns"`
			Areas             []area.Area   `yaml:"areas"`
		} `yaml:"terrorZone"`
		Leveling struct {
			EnsurePointsAllocation bool `yaml:"ensurePointsAllocation"`
			EnsureKeyBinding       bool `yaml:"ensureKeyBinding"`
		} `yaml:"leveling"`
	} `yaml:"game"`
	Companion struct {
		Enabled          bool   `yaml:"enabled"`
		Leader           bool   `yaml:"leader"`
		LeaderName       string `yaml:"leaderName"`
		Remote           string `yaml:"remote"`
		GameNameTemplate string `yaml:"gameNameTemplate"`
		GamePassword     string `yaml:"gamePassword"`
	} `yaml:"companion"`
	Gambling struct {
		Enabled bool        `yaml:"enabled"`
		Items   []item.Name `yaml:"items"`
	} `yaml:"gambling"`
	Runtime struct {
		CastDuration time.Duration
		Rules        []nip.Rule
	}
}

// Load reads the config.ini file and returns a Config struct filled with data from the ini file
func Load() error {
	Characters = make(map[string]*CharacterCfg)
	r, err := os.Open("config/koolo.yaml")
	if err != nil {
		return fmt.Errorf("error loading koolo.yaml: %w", err)
	}

	d := yaml.NewDecoder(r)
	if err = d.Decode(&Koolo); err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	entries, err := os.ReadDir("config")
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		charCfg := CharacterCfg{}
		r, err = os.Open("config/" + entry.Name() + "/config.yaml")
		if err != nil {
			return fmt.Errorf("error loading config.yaml: %w", err)
		}

		d := yaml.NewDecoder(r)
		if err = d.Decode(&charCfg); err != nil {
			return fmt.Errorf("error reading %s character config: %w", entry.Name(), err)
		}

		rules, err := nip.ReadDir("config/" + entry.Name() + "/pickit/")
		if err != nil {
			return err
		}

		if charCfg.Game.Runs[0] == "leveling" {
			levelingRules, err := nip.ReadDir("config/" + entry.Name() + "/pickit_leveling/")
			if err != nil {
				return err
			}
			rules = append(rules, levelingRules...)
		}

		charCfg.Runtime.Rules = rules

		secs := float32(charCfg.Character.CastingFrames)*0.04 + 0.01
		charCfg.Runtime.CastDuration = time.Duration(secs*1000) * time.Millisecond
		Characters[entry.Name()] = &charCfg
	}

	return nil
}

func CreateFromTemplate(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if _, err := os.Stat("config/" + name); !os.IsNotExist(err) {
		return errors.New("configuration with that name already exists")
	}

	err := cp.Copy("config/template", "config/"+name)
	if err != nil {
		return fmt.Errorf("error copying template: %w", err)
	}

	return Load()
}

func ValidateAndSaveConfig(config KooloCfg) error {
	// Trim executable from the path, just in case
	config.D2LoDPath = strings.ReplaceAll(strings.ToLower(config.D2LoDPath), "game.exe", "")
	config.D2RPath = strings.ReplaceAll(strings.ToLower(config.D2RPath), "d2r.exe", "")

	// Validate paths
	if _, err := os.Stat(config.D2LoDPath + "/d2data.mpq"); os.IsNotExist(err) {
		return errors.New("D2LoDPath is not valid")
	}

	if _, err := os.Stat(config.D2RPath + "/d2r.exe"); os.IsNotExist(err) {
		return errors.New("D2RPath is not valid")
	}

	text, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error parsing koolo config: %w", err)
	}

	err = os.WriteFile("config/koolo.yaml", text, 0644)
	if err != nil {
		return fmt.Errorf("error writing koolo config: %w", err)
	}

	return Load()
}
