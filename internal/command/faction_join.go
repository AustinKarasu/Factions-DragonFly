package command

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// FactionJoin defines the structure for the /f join <faction> command.
type FactionJoin struct {
	sessionManager *session.Manager
	Join           cmd.SubCommand `cmd:"join"`
	FactionName    string         `cmd:"faction"`
}

func (c FactionJoin) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}

	if _, ok := c.sessionManager.GetPlayerFaction(p.UUID()); ok {
		o.Errorf("§cYou are already in a faction.")
		return
	}

	invite, ok := c.sessionManager.GetInvitation(p.UUID())
	if !ok {
		o.Errorf("§cYou have no pending invitations.")
		return
	}
	invitingFaction, ok := c.sessionManager.GetFactionByID(invite.FactionID)
	if !ok {
		o.Errorf("§cThe faction that invited you no longer exists.")
		c.sessionManager.RemoveInvitation(p.UUID())
		return
	}
	if invitingFaction.Name != c.FactionName {
		o.Errorf("§cYou do not have an invitation to join '%s'.", c.FactionName)
		return
	}

	err := c.sessionManager.AddPlayerToFaction(p.UUID(), p.Name(), invitingFaction)
	if err != nil {
		o.Errorf("§cAn error occurred while joining the faction.")
		fmt.Printf("Error joining player %s to faction %s: %v\n", p.Name(), invitingFaction.Name, err)
		return
	}

	c.sessionManager.RemoveInvitation(p.UUID()) // This will stop the timer.
	o.Printf("§aYou have joined the faction '%s'.", invitingFaction.Name)
}
