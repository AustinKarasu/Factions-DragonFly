package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

// BountyOpen opens the bounty board UI.
type BountyOpen struct {
	sessionManager *session.Manager
}

func (c BountyOpen) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	openBountyMenu(p, c.sessionManager)
}

// BountyTop prints the top bounty targets.
type BountyTop struct {
	sessionManager *session.Manager
	Top            cmd.SubCommand `cmd:"top"`
}

func (c BountyTop) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	top := c.sessionManager.TopBounties(10)
	o.Printf("§6--- Top Bounties ---")
	for i, data := range top {
		if data.Bounty <= 0 {
			continue
		}
		o.Printf("§e#%d §f%s §7- §c$%d", i+1, data.Name, data.Bounty)
	}
}

// BountySet places a bounty on a player.
type BountySet struct {
	sessionManager *session.Manager
	Set            cmd.SubCommand `cmd:"set"`
	Target         []cmd.Target   `cmd:"player"`
	Amount         int            `cmd:"amount"`
}

func (c BountySet) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok || len(c.Target) != 1 {
		return
	}
	target, ok := c.Target[0].(*player.Player)
	if !ok {
		o.Errorf("§cTarget must be an online player.")
		return
	}
	if err := c.sessionManager.PlaceBounty(p.UUID(), target.UUID(), c.Amount); err != nil {
		o.Errorf("§cUnable to place bounty: %v", err)
		return
	}
	o.Printf("§aPlaced a $%d bounty on %s.", c.Amount, target.Name())
}

// BountyClear removes a bounty from a player.
type BountyClear struct {
	sessionManager *session.Manager
	Clear          cmd.SubCommand `cmd:"clear"`
	Target         []cmd.Target   `cmd:"player"`
}

func (c BountyClear) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	if len(c.Target) != 1 {
		o.Errorf("§cYou must choose one target.")
		return
	}
	target, ok := c.Target[0].(*player.Player)
	if !ok {
		o.Errorf("§cTarget must be online.")
		return
	}
	if err := c.sessionManager.ClearBounty(target.UUID()); err != nil {
		o.Errorf("§cUnable to clear bounty: %v", err)
		return
	}
	o.Printf("§aCleared %s's bounty.", target.Name())
}
