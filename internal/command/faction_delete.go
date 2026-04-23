package command

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// FactionDelete defines the structure for the /f delete <name> command.
type FactionDelete struct {
	sessionManager *session.Manager
	Delete         cmd.SubCommand `cmd:"delete"`
	FactionName    string         `cmd:"name"`
}

func (c FactionDelete) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}

	playerFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		replyErr(p, o, "§cYou are not in a faction.")
		return
	}
	if playerFaction.Leader != p.UUID() {
		replyErr(p, o, "§cOnly the leader can disband the faction.")
		return
	}
	if playerFaction.Name != c.FactionName {
		replyErr(p, o, "§cThe name does not match. Type '/f delete %s' to confirm.", playerFaction.Name)
		return
	}
	if err := c.sessionManager.DeleteFaction(playerFaction); err != nil {
		replyErr(p, o, "§cAn unexpected error occurred while disbanding the faction.")
		fmt.Printf("Error disbanding faction %s: %v\n", playerFaction.Name, err)
		return
	}

	reply(p, o, "§aYou successfully disbanded the faction '%s'.", playerFaction.Name)
	c.sessionManager.UpdateScoreboard(p)
}
