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
		replyErr(p, o, "§cYou are not in a faction.")
		return
	}
	if playerFaction.Leader == p.UUID() {
		replyErr(p, o, "§cThe leader cannot leave the faction. Promote someone or disband it with /f delete.")
		return
	}
	if err := c.sessionManager.RemovePlayerFromFaction(p.UUID(), playerFaction); err != nil {
		replyErr(p, o, "§cAn error occurred while leaving the faction.")
		return
	}

	reply(p, o, "§aYou left the faction '%s'.", playerFaction.Name)
	c.sessionManager.UpdateScoreboard(p)
}
