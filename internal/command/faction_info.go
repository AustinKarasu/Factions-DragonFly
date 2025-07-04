package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/faction"
	"github.com/jorgebyte/faction/internal/session"
	"strings"
)

// FactionInfoSelf defines the structure for the /f info command (no arguments).
type FactionInfoSelf struct {
	sessionManager *session.Manager
	Info           cmd.SubCommand `cmd:"info"`
}

// FactionInfoOther defines the structure for the /f info <name> command.
type FactionInfoOther struct {
	sessionManager *session.Manager
	Info           cmd.SubCommand `cmd:"info"`
	FactionName    string         `cmd:"name"`
}

func (c FactionInfoSelf) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}

	// Find the player's own faction.
	playerFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		o.Errorf("§cYou are not in a faction. Use /f info <name> to view another faction's information.")
		return
	}

	sendFactionInfo(playerFaction, o)
}

func (c FactionInfoOther) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	targetFaction, ok := c.sessionManager.GetFactionByName(c.FactionName)
	if !ok {
		o.Errorf("§cThe faction '%s' was not found.", c.FactionName)
		return
	}

	sendFactionInfo(targetFaction, o)
}

// sendFactionInfo is a helper function to avoid repeating code.
// It formats and sends a faction's information to the player.
func sendFactionInfo(f *faction.Faction, o *cmd.Output) {
	leaderName := f.Members[f.Leader]

	var coleaderNames []string
	for _, name := range f.Coleaders {
		coleaderNames = append(coleaderNames, name)
	}

	var memberNames []string
	for memberID, memberName := range f.Members {
		// Only add to the list if they are not the leader AND not a co-leader.
		if memberID != f.Leader {
			if _, isColeader := f.Coleaders[memberID]; !isColeader {
				memberNames = append(memberNames, memberName)
			}
		}
	}

	var allyNames []string
	for _, allyName := range f.Allies {
		allyNames = append(allyNames, allyName)
	}

	alliesText := "None"
	if len(allyNames) > 0 {
		alliesText = strings.Join(allyNames, ", ")
	}

	coleadersText := "None"
	if len(coleaderNames) > 0 {
		coleadersText = strings.Join(coleaderNames, ", ")
	}
	membersText := "None"
	if len(memberNames) > 0 {
		membersText = strings.Join(memberNames, ", ")
	}

	o.Printf("§e--- Information for %s §e---", f.Name)
	o.Printf("§bLeader: §f%s", leaderName)
	o.Printf("§bCo-Leaders: §f%s", coleadersText)
	o.Printf("§bAllies: §a%s", alliesText)
	o.Printf("§bMembers (%d): §f%s", len(f.Members), membersText)
	o.Printf("§bPower: §f%d", f.Power)
	o.Printf("§bCreated On: §f%s", f.CreatedAt.Format("01/02/2006"))
}
