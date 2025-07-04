package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/chunk"
	"github.com/jorgebyte/faction/internal/session"
)

// FactionUnclaim defines the structure for the /f unclaim command.
type FactionUnclaim struct {
	sessionManager *session.Manager
	Unclaim        cmd.SubCommand `cmd:"unclaim"`
}

func (c FactionUnclaim) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
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
		o.Errorf("§cOnly the faction leader can unclaim territories.")
		return
	}

	currentChunk := chunk.FromWorldPos(p.Position())
	if err := c.sessionManager.UnclaimChunk(currentChunk, playerFaction); err != nil {
		o.Errorf("§cError unclaiming territory: %v", err)
		return
	}

	o.Printf("§aTerritory at %s has been unclaimed.", currentChunk.String())
}
