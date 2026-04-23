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
	handleFactionClaim(p, o, c.sessionManager)
}

func handleFactionClaim(p *player.Player, o *cmd.Output, sessionManager *session.Manager) {
	playerFaction, ok := sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		if o != nil {
			o.Errorf("§cYou are not in a faction to claim land.")
		}
		p.Message("§cYou are not in a faction to claim land.")
		return
	}

	if playerFaction.Leader != p.UUID() && playerFaction.Coleaders[p.UUID()] == "" {
		if o != nil {
			o.Errorf("§cOnly the faction leader or co-leaders can claim new territories.")
		}
		p.Message("§cOnly the faction leader or co-leaders can claim new territories.")
		return
	}

	currentChunk := chunk.FromWorldPos(p.Position())
	if err := sessionManager.ClaimChunk(currentChunk, playerFaction); err != nil {
		if o != nil {
			o.Errorf("§cError claiming territory: %v", err)
		}
		p.Messagef("§cError claiming territory: %v", err)
		return
	}

	msg := "§aTerritory claimed for '" + playerFaction.Name + "' at " + currentChunk.String() + "."
	if o != nil {
		o.Printf(msg)
	}
	p.Message(msg)
	sessionManager.UpdateScoreboard(p)
}
