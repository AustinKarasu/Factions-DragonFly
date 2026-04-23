package command

import (
	"time"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// FactionInvite defines the structure for the /f invite <player> command.
type FactionInvite struct {
	sessionManager *session.Manager
	Invite         cmd.SubCommand `cmd:"invite"`
	Target         []cmd.Target   `cmd:"player"`
}

func (c FactionInvite) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}

	inviterFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		replyErr(p, o, "§cYou are not in a faction to invite players.")
		return
	}
	if inviterFaction.Leader != p.UUID() && inviterFaction.Coleaders[p.UUID()] == "" {
		replyErr(p, o, "§cOnly the leader and co-leaders can invite players.")
		return
	}
	if len(c.Target) != 1 {
		replyErr(p, o, "§cYou must specify a single player.")
		return
	}

	targetPlayer, ok := c.Target[0].(*player.Player)
	if !ok {
		replyErr(p, o, "§cInvalid player.")
		return
	}
	if _, ok := c.sessionManager.GetPlayerFaction(targetPlayer.UUID()); ok {
		replyErr(p, o, "§cPlayer '%s' already belongs to a faction.", targetPlayer.Name())
		return
	}

	c.sessionManager.AddInvitation(inviterFaction.ID, targetPlayer.UUID(), 5*time.Minute)
	reply(p, o, "§aYou invited '%s' to your faction. The invitation expires in 5 minutes.", targetPlayer.Name())
	targetPlayer.Messagef("§eYou were invited to the faction '%s'. Use §a/f join %s §eto accept.", inviterFaction.Name, inviterFaction.Name)
}
