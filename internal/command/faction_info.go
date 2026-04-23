package command

import (
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/faction"
	"github.com/jorgebyte/faction/internal/session"
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
	playerFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		o.Errorf("§cYou are not in a faction. Use /f info <name> to view another faction.")
		return
	}
	sendFactionInfo(playerFaction, o, nil)
}

func (c FactionInfoOther) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	targetFaction, ok := c.sessionManager.GetFactionByName(c.FactionName)
	if !ok {
		o.Errorf("§cThe faction '%s' was not found.", c.FactionName)
		return
	}
	sendFactionInfo(targetFaction, o, nil)
}

// sendFactionInfo is a helper function to avoid repeating code.
func sendFactionInfo(f *faction.Faction, o *cmd.Output, p *player.Player) {
	leaderName := f.Members[f.Leader]

	var coleaderNames []string
	for _, name := range f.Coleaders {
		coleaderNames = append(coleaderNames, name)
	}

	var memberNames []string
	for memberID, memberName := range f.Members {
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

	lines := []string{
		"§e--- Information for " + f.Name + " §e---",
		"§bLeader: §f" + leaderName,
		"§bCo-Leaders: §f" + coleadersText,
		"§bAllies: §a" + alliesText,
		"§bMembers (" + strconv.Itoa(len(f.Members)) + "): §f" + membersText,
		"§bPower: §f" + strconv.Itoa(f.Power),
		"§bClaims: §f" + strconv.Itoa(f.Claims) + "/" + strconv.Itoa(f.ClaimLimit()),
		"§bDescription: §f" + f.Description,
		"§bCreated On: §f" + f.CreatedAt.Format("01/02/2006"),
	}

	for _, line := range lines {
		if o != nil {
			o.Printf(line)
		}
		if p != nil {
			p.Message(line)
		}
	}
}
