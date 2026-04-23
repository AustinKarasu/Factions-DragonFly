package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// Balance shows the player's current balance.
type Balance struct {
	sessionManager *session.Manager
}

func (c Balance) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	data := c.sessionManager.GetOrCreatePlayer(p.UUID(), p.Name())
	o.Printf("§aYour balance: §f$%d", data.Balance)
}

// Pay transfers money to another online player.
type Pay struct {
	sessionManager *session.Manager
	Target         []cmd.Target `cmd:"player"`
	Amount         int          `cmd:"amount"`
}

func (c Pay) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok || len(c.Target) != 1 {
		return
	}
	target, ok := c.Target[0].(*player.Player)
	if !ok {
		o.Errorf("§cTarget must be online.")
		return
	}
	if c.Amount <= 0 {
		o.Errorf("§cAmount must be positive.")
		return
	}
	if err := c.sessionManager.DebitBalance(p.UUID(), c.Amount); err != nil {
		o.Errorf("§cUnable to pay: %v", err)
		return
	}
	if err := c.sessionManager.CreditBalance(target.UUID(), c.Amount); err != nil {
		_ = c.sessionManager.CreditBalance(p.UUID(), c.Amount)
		o.Errorf("§cUnable to pay that player right now.")
		return
	}
	c.sessionManager.UpdateScoreboard(p)
	c.sessionManager.UpdateScoreboard(target)
	o.Printf("§aPaid $%d to %s.", c.Amount, target.Name())
	target.Messagef("§aYou received $%d from %s.", c.Amount, p.Name())
}
