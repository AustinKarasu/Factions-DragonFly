package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// FactionBorder defines the structure for the /f border command.
type FactionBorder struct {
	sessionManager *session.Manager
	Border         cmd.SubCommand `cmd:"border"`
}

func (c FactionBorder) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}

	if newState := c.sessionManager.ToggleBorderView(p.UUID()); newState {
		o.Printf("§aTerritory border view enabled.")
	} else {
		o.Printf("§eTerritory border view disabled.")
	}
}
