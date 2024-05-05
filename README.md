<p align="center">
  <img src="assets/koolo.webp" alt="Koolo" width="150">
</p>
<h3 align="center">Koolo</h3>

---

Koolo is a small bot for Diablo II: Resurrected. Koolo project was built for informational and educational purposes
only, it's not intended for online usage. Feel free to contribute opening pull requests with new features or bugfixes.
Koolo reads game memory and interacts with the game injecting clicks/keystrokes to the game window. As good as it can.

## Disclaimer
Can I get banned for using Koolo? The answer is a crystal clear yes, you can get banned although at this point I'm
not aware of any ban for using it. I'm not responsible for any ban or any other consequence that may arise from it.

## Features
- Blizzard Sorceress, Nova Sorceress and Hammerdin are currently supported
- Supported runs: Countess, Andariel, Ancient Tunnels, Summoner, Mephisto, Council, Eldritch, Pindleskin, Nihlathak,
  Tristram, Lower Kurast, Stony Tomb, The Pit, Arachnid Lair, Baal, Tal Rasha Tombs, Diablo (WIP), Cows
- Bot integration for Discord and Telegram
- "Companion mode" one leader bot will be creating games and the rest of the bots will join the game... and sometimes it
  works
- Pickit based on NIP files
- Auto potion for health and mana (also mercenary)
- Chicken when low health
- Inventory slot locking
- Revive mercenary
- CTA buff and class buffs
- Auto repair
- Skip on immune
- Auto leveling sorceress and paladin (WIP)
- Auto gambling

## Requirements
- Diablo II: Resurrected (1280x720 required, windowed mode, ensure accessibility large fonts disabled)
- **Diablo II: LOD 1.13c** (IMPORTANT: It will **NOT** work without it, this step is not optional)

## Quick Start
- If you haven't done yet, install **Diablo II: LOD 1.13c** (required)
- [Download](https://github.com/hectorgimenez/koolo/releases) the latest Koolo release (recommended for most users), or alternatively you can [build it from source](#development-environment)
- Extract the zip file in a directory of your choice.
- Run `koolo.exe`.
- Follow the setup wizard, it will guide you through the process of setting up the bot, you will need to setup some directories and character configuration.
- If you want to back up/restore your configuration, and for manual setup, you can find the configuration files in the `config` directory.

## Pickit rules
Item pickit is based on [NIP files](https://github.com/blizzhackers/pickits/blob/master/NipGuide.md), you can find them in the `config/{character}/pickit` directory.

All the .nip files contained in the pickit directory will be loaded, so you can have multiple pickit files.

There are some considerations to take into account:
- If item fully matches the pickit rule before being identified, it will be picked up and stashed unidentified.
- If item doesn't match the full rule, will be identified and checked again, if fully matches a rule it will be stashed.
- There are some NIP properties that are **not implemented yet** and the bot will let you know during the startup process, please be sure to not use them:
  - plusmindamage
  - mindamage (Only working for items with no base damage like rings, amulets, armors, etc.)
  - plusmaxdamage
  - maxdamage (Only working for items with no base damage like rings, amulets, armors, etc.)
  - enhanceddamage (Only working for items with no base damage like rings, amulets, armors, etc.)
  - enhanceddefense (Only working for white bases and is not 100% accurate, avoid using it as much as possible)
  - itemarmorpercent
  - itemmindamagepercent
  - itemslashdamage
  - itemslashdamagepercent
  - itemcrushdamage
  - itemcrushdamagepercent
  - itemthrustdamage
  - itemthrustdamagepercent
  - secondarymindamage
  - secondarymaxdamage
  - damagepercent

## Development environment
**Note:** This is only required if you want to build the project from source. If you want to run the bot, you can just download the [latest release](https://github.com/hectorgimenez/koolo/releases).

Setting the development environment is pretty straightforward, but the following dependencies are **required** to build the project.

### Dependencies
- [Download Go >= 1.22](https://go.dev/dl/)
- [Install git](https://gitforwindows.org/)

### Building from source
Open the terminal and run the following commands in project root directory:
```shell
git clone https://github.com/hectorgimenez/koolo.git
cd koolo
build.bat
```
This will produce the "build" directory with the executable file and all the required assets.

### Updating with latest changes
In order to fetch latest `main` branch changes run the following commands in project root directory:
```shell
git pull
build.bat
```
**Note**: `build` directory **will be deleted**, so if you customized any file in there, make sure to backup it before running `build.bat`.
