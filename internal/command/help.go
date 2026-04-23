package command

import (
	"fmt"
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/chunk"
	"github.com/jorgebyte/faction/internal/session"
)

type FactionHelp struct {
	sessionManager *session.Manager
	Help           cmd.SubCommand `cmd:"help"`
}

type FactionOverview struct {
	sessionManager *session.Manager
	Overview       cmd.SubCommand `cmd:"overview"`
}

type FactionWho struct {
	sessionManager *session.Manager
	Who            cmd.SubCommand `cmd:"who"`
}

type FactionMap struct {
	sessionManager *session.Manager
	Map            cmd.SubCommand `cmd:"map"`
}

func (c FactionHelp) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	lines := factionHelpLines()
	for _, line := range lines {
		o.Printf(line)
	}
	if p, ok := src.(*player.Player); ok {
		p.SendForm(form.NewMenu(menuSubmit{}, "Faction Help").WithBody(stringsJoin(lines)).WithButtons(
			form.NewButton("Close", "textures/ui/cancel"),
		))
	}
}

func (c FactionOverview) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	playerFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		o.Errorf("§cYou are not in a faction.")
		return
	}
	sendFactionInfo(playerFaction, o, p)
}

func (c FactionWho) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	playerFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		o.Errorf("§cYou are not in a faction.")
		return
	}
	lines := []string{fmt.Sprintf("§e%s members (%d)", playerFaction.Name, len(playerFaction.Members))}
	for memberID, name := range playerFaction.Members {
		role := "Member"
		if memberID == playerFaction.Leader {
			role = "Leader"
		} else if _, ok := playerFaction.Coleaders[memberID]; ok {
			role = "Co-Leader"
		}
		status := "§7Offline"
		if _, ok := c.sessionManager.GetOnlinePlayerByName(name); ok {
			status = "§aOnline"
		}
		lines = append(lines, fmt.Sprintf("§f%s §7- %s §7(%s§7)", name, role, status))
	}
	for _, line := range lines {
		o.Printf(line)
	}
	p.SendForm(form.NewMenu(menuSubmit{}, "Faction Members").WithBody(stringsJoin(lines)).WithButtons(
		form.NewButton("Close", "textures/ui/cancel"),
	))
}

func (c FactionMap) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	current := chunk.FromWorldPos(p.Position())
	lines := []string{"§eClaim map legend: §aYours §cClaimed §7Wilderness"}
	for z := current.Z - 1; z <= current.Z+1; z++ {
		row := make([]string, 0, 3)
		for x := current.X - 1; x <= current.X+1; x++ {
			pos := chunk.Pos{X: x, Z: z}
			marker := "§7[ ]"
			if ownerID, ok := c.sessionManager.GetClaimOwner(pos); ok {
				if fac, ok := c.sessionManager.GetPlayerFaction(p.UUID()); ok && fac.ID == ownerID {
					marker = "§a[X]"
				} else {
					marker = "§c[X]"
				}
			}
			if pos == current {
				marker = strings.Replace(marker, "[", "{", 1)
				marker = strings.Replace(marker, "]", "}", 1)
			}
			row = append(row, marker)
		}
		lines = append(lines, strings.Join(row, " "))
	}
	lines = append(lines, fmt.Sprintf("§fCurrent chunk: §b%s", current.String()))
	for _, line := range lines {
		o.Printf(line)
	}
	p.SendForm(form.NewMenu(menuSubmit{}, "Faction Map").WithBody(stringsJoin(lines)).WithButtons(
		form.NewButton("Close", "textures/ui/cancel"),
	))
}

func factionHelpLines() []string {
	return []string{
		"§eFaction commands",
		"§f/f §7- open the faction control menu",
		"§f/f help §7- show faction help and usage",
		"§f/f create <name> §7- create a faction",
		"§f/f overview §7- show your faction details",
		"§f/f who §7- list faction members and online status",
		"§f/f invite <player> §7- invite a player",
		"§f/f join <faction> §7- accept a faction invite",
		"§f/f leave §7- leave your faction",
		"§f/f claim §7- claim the current chunk",
		"§f/f unclaim §7- unclaim the current chunk",
		"§f/f border §7- toggle chunk border particles",
		"§f/f map §7- show a small claim map",
		"§f/f top faction §7- top factions by power",
		"§f/f top player §7- top players by power",
		"§f/bounty §7- open the bounty board",
		"§f/shop §7- open the faction shop",
		"§f/balance, /pay §7- economy commands",
	}
}
