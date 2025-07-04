package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/chunk"
	"github.com/jorgebyte/faction/internal/session"
)

// FactionClaim defines the structure for the /f claim command.
type FactionClaim struct {
	sessionManager *session.Manager
	Claim          cmd.SubCommand `cmd:"claim"`
}

func (c FactionClaim) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}

	playerFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		o.Errorf("§cYou are not in a faction to claim land.")
		return
	}

	if playerFaction.Leader != p.UUID() {
		o.Errorf("§cOnly the faction leader can claim new territories.")
		return
	}

	currentChunk := chunk.FromWorldPos(p.Position())
	if err := c.sessionManager.ClaimChunk(currentChunk, playerFaction); err != nil {
		o.Errorf("§cError claiming territory: %v", err)
		return
	}

	o.Printf("§aTerritory claimed for '%s' at %s.", playerFaction.Name, currentChunk.String())
}
