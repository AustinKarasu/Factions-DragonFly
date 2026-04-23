package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// FactionMenu opens the main UI for /f.
type FactionMenu struct {
	sessionManager *session.Manager
}

func (c FactionMenu) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	openFactionMainMenu(p, c.sessionManager)
}
