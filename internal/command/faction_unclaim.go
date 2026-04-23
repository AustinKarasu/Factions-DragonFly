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
	handleFactionUnclaim(p, o, c.sessionManager)
}

func handleFactionUnclaim(p *player.Player, o *cmd.Output, sessionManager *session.Manager) {
	playerFaction, ok := sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		replyErr(p, o, "§cYou are not in a faction.")
		return
	}

	if playerFaction.Leader != p.UUID() && playerFaction.Coleaders[p.UUID()] == "" {
		replyErr(p, o, "§cOnly the faction leader or co-leaders can unclaim territories.")
		return
	}

	currentChunk := chunk.FromWorldPos(p.Position())
	if err := sessionManager.UnclaimChunk(currentChunk, playerFaction); err != nil {
		replyErr(p, o, "§cError unclaiming territory: %v", err)
		return
	}

	reply(p, o, "§aTerritory at %s has been unclaimed.", currentChunk.String())
	sessionManager.UpdateScoreboard(p)
}
