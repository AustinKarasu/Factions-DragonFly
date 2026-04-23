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
		replyErr(p, o, "§cYou are already in a faction. Use /f leave to exit.")
		return
	}
	if _, ok := c.sessionManager.GetFactionByName(c.FactionName); ok {
		replyErr(p, o, "§cThe faction name '%s' already exists.", c.FactionName)
		return
	}

	newFaction := faction.New(c.FactionName, p.UUID(), p.Name())
	if err := c.sessionManager.CreateFaction(newFaction); err != nil {
		replyErr(p, o, "§cAn unexpected error occurred while creating the faction.")
		fmt.Printf("Error creating faction %s: %v\n", c.FactionName, err)
		return
	}

	reply(p, o, "§aCongratulations! You created the faction '%s'.", c.FactionName)
	c.sessionManager.UpdateScoreboard(p)
}
