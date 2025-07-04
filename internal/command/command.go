package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// SubCommand defines the structure that all of our subcommands must have.
type SubCommand interface {
	// Name is the name of the subcommand (e.g., "create", "invite").
	Name() string
	// Execute is the function called when a player uses the subcommand.
	Execute(p *player.Player, args []string, s *session.Manager)
}

// Base is a base command that can hold multiple subcommands.
type Base struct {
	Name           string
	Description    string
	Aliases        []string
	SubCommands    map[string]SubCommand
	SessionManager *session.Manager
}

func (b *Base) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	panic("implement me")
}

func (b *Base) Execute(src cmd.Source, args []string) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}

	if len(args) == 0 {
		p.Message("§eUse /f help for a list of commands.")
		return
	}

	subCmd, ok := b.SubCommands[args[0]]
	if !ok {
		p.Messagef("§cThe subcommand '%s' does not exist.", args[0])
		return
	}

	subCmd.Execute(p, args[1:], b.SessionManager)
}
