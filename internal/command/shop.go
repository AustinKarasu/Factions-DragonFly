package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// ShopOpen opens the main shop UI.
type ShopOpen struct {
	sessionManager *session.Manager
}

func (c ShopOpen) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	openShopCategoriesMenu(p, c.sessionManager)
}

// ShopAdmin opens the admin UI for shop editing.
type ShopAdmin struct {
	sessionManager *session.Manager
	Admin          cmd.SubCommand `cmd:"admin"`
}

func (c ShopAdmin) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	openShopAdminMenu(p, c.sessionManager)
}
