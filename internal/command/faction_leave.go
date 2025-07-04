package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// FactionLeave defines the structure for the /f leave command.
type FactionLeave struct {
	sessionManager *session.Manager
	Leave          cmd.SubCommand `cmd:"leave"`
}

func (c FactionLeave) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}

	playerFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		o.Errorf("§cYou are not in a faction.")
		return
	}

	if playerFaction.Leader == p.UUID() {
		o.Errorf("§cThe leader cannot leave the faction. You must appoint a new leader or disband it with /f delete.")
		return
	}

	if err := c.sessionManager.RemovePlayerFromFaction(p.UUID(), playerFaction); err != nil {
		o.Errorf("§cAn error occurred while leaving the faction.")
		return
	}
	o.Printf("§aYou have left the faction '%s'.", playerFaction.Name)
}
