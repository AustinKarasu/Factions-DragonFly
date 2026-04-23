package command

import (
	"fmt"
	"slices"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/faction"
	playerdata "github.com/jorgebyte/faction/internal/player"
	"github.com/jorgebyte/faction/internal/session"
)

type TopType string

func (TopType) Type() string {
	return "TopType"
}

func (TopType) Options(src cmd.Source) []string {
	return []string{"player", "players", "faction", "factions"}
}

type FactionTop struct {
	sessionManager *session.Manager
	Top            cmd.SubCommand `cmd:"top"`
	Category       TopType        `cmd:"category"`
}

func (c FactionTop) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	switch string(c.Category) {
	case "player", "players":
		if p, ok := src.(*player.Player); ok {
			openTopPlayersMenu(p, c.sessionManager)
			return
		}
		c.showPlayerTop(o)
	case "faction", "factions":
		if p, ok := src.(*player.Player); ok {
			openTopFactionsMenu(p, c.sessionManager)
			return
		}
		c.showFactionTop(o)
	}
}

func (c FactionTop) showPlayerTop(o *cmd.Output) {
	allPlayers := c.sessionManager.GetAllPlayersData()
	slices.SortFunc(allPlayers, func(a, b *playerdata.Data) int {
		return b.Power - a.Power
	})

	o.Printf("§e--- Top 10 Players by Power ---")
	limit := min(10, len(allPlayers))
	if limit == 0 {
		o.Printf("§cThere are no players in the ranking yet.")
		return
	}
	for i := 0; i < limit; i++ {
		pData := allPlayers[i]
		o.Printf("§6#%d §f%s §7- §b%d Power", i+1, pData.Name, pData.Power)
	}
}

func (c FactionTop) showFactionTop(o *cmd.Output) {
	allFactions := c.sessionManager.GetAllFactionsData()
	allPlayers := c.sessionManager.GetAllPlayersMap()

	slices.SortFunc(allFactions, func(a, b *faction.Faction) int {
		return b.CalculatePower(allPlayers) - a.CalculatePower(allPlayers)
	})

	o.Printf("§e--- Top 10 Factions by Power ---")
	limit := min(10, len(allFactions))
	if limit == 0 {
		o.Printf("§cThere are no factions in the ranking yet.")
		return
	}
	for i := 0; i < limit; i++ {
		f := allFactions[i]
		totalPower := f.CalculatePower(allPlayers)
		o.Printf("§6#%d §f%s §7- §b%d Power §7(%d claims)", i+1, f.Name, totalPower, f.Claims)
	}
}

func openTopPlayersMenu(p *player.Player, s *session.Manager) {
	allPlayers := s.GetAllPlayersData()
	slices.SortFunc(allPlayers, func(a, b *playerdata.Data) int {
		return b.Power - a.Power
	})

	lines := []string{"§eTop players by power"}
	limit := min(10, len(allPlayers))
	if limit == 0 {
		lines = append(lines, "§cNo players are ranked yet.")
	} else {
		for i := 0; i < limit; i++ {
			data := allPlayers[i]
			lines = append(lines, fmt.Sprintf("§6#%d §f%s §7- §b%d power", i+1, data.Name, data.Power))
		}
	}
	p.SendForm(form.NewMenu(menuSubmit{}, "Top Players").WithBody(stringsJoin(lines)).WithButtons(
		form.NewButton("Close", "textures/ui/cancel"),
	))
}

func openTopFactionsMenu(p *player.Player, s *session.Manager) {
	allFactions := s.GetAllFactionsData()
	allPlayers := s.GetAllPlayersMap()
	slices.SortFunc(allFactions, func(a, b *faction.Faction) int {
		return b.CalculatePower(allPlayers) - a.CalculatePower(allPlayers)
	})

	lines := []string{"§eTop factions by power"}
	limit := min(10, len(allFactions))
	if limit == 0 {
		lines = append(lines, "§cNo factions are ranked yet.")
	} else {
		for i := 0; i < limit; i++ {
			f := allFactions[i]
			lines = append(lines, fmt.Sprintf("§6#%d §f%s §7- §b%d power §7(%d claims)", i+1, f.Name, f.CalculatePower(allPlayers), f.Claims))
		}
	}
	p.SendForm(form.NewMenu(menuSubmit{}, "Top Factions").WithBody(stringsJoin(lines)).WithButtons(
		form.NewButton("Close", "textures/ui/cancel"),
	))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
