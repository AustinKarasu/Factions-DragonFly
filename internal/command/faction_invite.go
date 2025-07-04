package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
	"time"
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
		o.Errorf("§cYou are not in a faction to invite players.")
		return
	}

	if inviterFaction.Leader != p.UUID() {
		o.Errorf("§cOnly the leader and co-leaders can invite players.")
		return
	}

	if len(c.Target) != 1 {
		o.Errorf("§cYou must specify a single player.")
		return
	}

	targetPlayer, ok := c.Target[0].(*player.Player)
	if !ok {
		o.Errorf("§cInvalid player.")
		return
	}
	if _, ok := c.sessionManager.GetPlayerFaction(targetPlayer.UUID()); ok {
		o.Errorf("§cPlayer '%s' already belongs to a faction.", targetPlayer.Name())
		return
	}

	expirationTime := 5 * time.Minute // The invitation will last for 5 minutes.
	c.sessionManager.AddInvitation(inviterFaction.ID, targetPlayer.UUID(), expirationTime)

	o.Printf("§aYou have invited '%s' to your faction. The invitation expires in 5 minutes.", targetPlayer.Name())
	targetPlayer.Messagef("§eYou have been invited to the faction '%s'. Use §a/f join %s §eto accept. The invitation expires in 5 minutes!", inviterFaction.Name, inviterFaction.Name)
}
