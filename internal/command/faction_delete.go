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
		o.Errorf("§cYou are not in a faction.")
		return
	}

	if playerFaction.Leader != p.UUID() {
		o.Errorf("§cOnly the leader can disband the faction.")
		return
	}

	if playerFaction.Name != c.FactionName {
		o.Errorf("§cThe name does not match. Type '/f delete %s' to confirm.", playerFaction.Name)
		return
	}

	err := c.sessionManager.DeleteFaction(playerFaction)
	if err != nil {
		o.Errorf("§cAn unexpected error occurred while disbanding the faction.")
		fmt.Printf("Error disbanding faction %s: %v\n", playerFaction.Name, err)
		return
	}

	o.Printf("§aYou have successfully disbanded the faction '%s'.", playerFaction.Name)
}
