package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/faction"
	"github.com/jorgebyte/faction/internal/player"
	"github.com/jorgebyte/faction/internal/session"
	"slices"
)

// TopType is an Enum for the player to choose which ranking to view.
type TopType string

// Type returns the name of the Enum type.
func (TopType) Type() string {
	return "TopType"
}

// Options returns the valid options for the Enum.
func (TopType) Options(src cmd.Source) []string {
	return []string{"player", "faction"}
}

// FactionTop defines the structure for the /f top <player|faction> command.
type FactionTop struct {
	sessionManager *session.Manager
	Top            cmd.SubCommand `cmd:"top"`
	Category       TopType        `cmd:"category"`
}

func (c FactionTop) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	switch c.Category {
	case "player":
		c.showPlayerTop(o)
	case "faction":
		c.showFactionTop(o)
	}
}

// showPlayerTop displays the ranking of the top 10 players.
func (c FactionTop) showPlayerTop(o *cmd.Output) {
	allPlayers := c.sessionManager.GetAllPlayersData()
	slices.SortFunc(allPlayers, func(a, b *player.Data) int {
		return b.Power - a.Power
	})

	o.Printf("§e--- Top 10 Players by Power ---")
	limit := 10
	if len(allPlayers) < limit {
		limit = len(allPlayers)
	}
	if limit == 0 {
		o.Printf("§cThere are no players in the ranking yet.")
		return
	}
	for i := 0; i < limit; i++ {
		pData := allPlayers[i]
		o.Printf("§6#%d §f%s §7- §b%d Power", i+1, pData.Name, pData.Power)
	}
}

// showFactionTop displays the ranking of the top 10 factions.
func (c FactionTop) showFactionTop(o *cmd.Output) {
	allFactions := c.sessionManager.GetAllFactionsData()
	allPlayers := c.sessionManager.GetAllPlayersMap() // Well need this method.

	slices.SortFunc(allFactions, func(a, b *faction.Faction) int {
		powerA := a.CalculatePower(allPlayers)
		powerB := b.CalculatePower(allPlayers)
		return powerB - powerA
	})

	o.Printf("§e--- Top 10 Factions by Power ---")
	limit := 10
	if len(allFactions) < limit {
		limit = len(allFactions)
	}
	if limit == 0 {
		o.Printf("§cThere are no factions in the ranking yet.")
		return
	}
	for i := 0; i < limit; i++ {
		f := allFactions[i]
		totalPower := f.CalculatePower(allPlayers)
		o.Printf("§6#%d §f%s §7- §b%d Power", i+1, f.Name, totalPower)
	}
}
