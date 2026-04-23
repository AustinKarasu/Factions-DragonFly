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
		if o != nil {
			o.Errorf("§cYou are not in a faction.")
		}
		p.Message("§cYou are not in a faction.")
		return
	}

	if playerFaction.Leader != p.UUID() && playerFaction.Coleaders[p.UUID()] == "" {
		if o != nil {
			o.Errorf("§cOnly the faction leader or co-leaders can unclaim territories.")
		}
		p.Message("§cOnly the faction leader or co-leaders can unclaim territories.")
		return
	}

	currentChunk := chunk.FromWorldPos(p.Position())
	if err := sessionManager.UnclaimChunk(currentChunk, playerFaction); err != nil {
		if o != nil {
			o.Errorf("§cError unclaiming territory: %v", err)
		}
		p.Messagef("§cError unclaiming territory: %v", err)
		return
	}

	msg := "§aTerritory at " + currentChunk.String() + " has been unclaimed."
	if o != nil {
		o.Printf(msg)
	}
	p.Message(msg)
	sessionManager.UpdateScoreboard(p)
}
