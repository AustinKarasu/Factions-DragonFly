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
		replyErr(p, o, "§cYou are not in a faction to claim land.")
		return
	}

	if playerFaction.Leader != p.UUID() && playerFaction.Coleaders[p.UUID()] == "" {
		replyErr(p, o, "§cOnly the faction leader or co-leaders can claim new territories.")
		return
	}

	currentChunk := chunk.FromWorldPos(p.Position())
	if err := sessionManager.ClaimChunk(currentChunk, playerFaction); err != nil {
		replyErr(p, o, "§cError claiming territory: %v", err)
		return
	}

	reply(p, o, "§aTerritory claimed for '%s' at %s.", playerFaction.Name, currentChunk.String())
	sessionManager.UpdateScoreboard(p)
}
