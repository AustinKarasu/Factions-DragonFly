package command

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/faction"
	"github.com/jorgebyte/faction/internal/session"
)

// FactionCreate defines the structure for the /f create <name> command.
type FactionCreate struct {
	sessionManager *session.Manager
	Create         cmd.SubCommand `cmd:"create"`
	FactionName    string         `cmd:"name"`
}

func (c FactionCreate) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}

	if _, ok := c.sessionManager.GetPlayerFaction(p.UUID()); ok {
		o.Errorf("§cYou are already in a faction. Use /f leave to exit.")
		return
	}

	if _, ok := c.sessionManager.GetFactionByName(c.FactionName); ok {
		o.Errorf("§cThe faction name '%s' already exists.", c.FactionName)
		return
	}

	newFaction := faction.New(c.FactionName, p.UUID(), p.Name())
	err := c.sessionManager.CreateFaction(newFaction)
	if err != nil {
		o.Errorf("§cAn unexpected error occurred while creating the faction. Contact an administrator.")
		fmt.Printf("Error creating faction %s: %v\n", c.FactionName, err)
		return
	}

	o.Printf("§aCongratulations! You have created the faction '%s'.", c.FactionName)
}
