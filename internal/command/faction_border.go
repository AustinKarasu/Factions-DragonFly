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
	handleFactionBorderToggle(p, o, c.sessionManager)
}

func handleFactionBorderToggle(p *player.Player, o *cmd.Output, sessionManager *session.Manager) {
	if newState := sessionManager.ToggleBorderView(p.UUID()); newState {
		if o != nil {
			o.Printf("§aTerritory border view enabled.")
		}
		p.Message("§aTerritory border view enabled.")
	} else {
		if o != nil {
			o.Printf("§eTerritory border view disabled.")
		}
		p.Message("§eTerritory border view disabled.")
	}
}
