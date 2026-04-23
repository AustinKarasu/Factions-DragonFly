package command

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/session"
)

type FactionDesc struct {
	sessionManager *session.Manager
	Desc           cmd.SubCommand `cmd:"desc"`
	Description    string         `cmd:"description"`
}

type FactionKick struct {
	sessionManager *session.Manager
	Kick           cmd.SubCommand `cmd:"kick"`
	Target         []cmd.Target   `cmd:"player"`
}

type FactionPromote struct {
	sessionManager *session.Manager
	Promote        cmd.SubCommand `cmd:"promote"`
	Target         []cmd.Target   `cmd:"player"`
}

type FactionDemote struct {
	sessionManager *session.Manager
	Demote         cmd.SubCommand `cmd:"demote"`
	Target         []cmd.Target   `cmd:"player"`
}

func (c FactionDesc) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	fac, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		replyErr(p, o, "§cYou are not in a faction.")
		return
	}
	if fac.Leader != p.UUID() && fac.Coleaders[p.UUID()] == "" {
		replyErr(p, o, "§cOnly the leader or co-leaders can update the faction description.")
		return
	}
	description := strings.TrimSpace(c.Description)
	if description == "" {
		replyErr(p, o, "§cDescription cannot be empty.")
		return
	}
	fac.Description = description
	if err := c.sessionManager.SaveFaction(fac); err != nil {
		replyErr(p, o, "§cUnable to update faction description right now.")
		return
	}
	reply(p, o, "§aFaction description updated.")
}

func (c FactionKick) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	fac, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		replyErr(p, o, "§cYou are not in a faction.")
		return
	}
	if fac.Leader != p.UUID() && fac.Coleaders[p.UUID()] == "" {
		replyErr(p, o, "§cOnly the leader or co-leaders can kick members.")
		return
	}
	if len(c.Target) != 1 {
		replyErr(p, o, "§cChoose exactly one online player.")
		return
	}
	target, ok := c.Target[0].(*player.Player)
	if !ok {
		replyErr(p, o, "§cThat target is invalid.")
		return
	}
	if target.UUID() == p.UUID() {
		replyErr(p, o, "§cUse /f leave if you want to leave your faction.")
		return
	}
	targetFaction, ok := c.sessionManager.GetPlayerFaction(target.UUID())
	if !ok || targetFaction.ID != fac.ID {
		replyErr(p, o, "§cThat player is not in your faction.")
		return
	}
	if target.UUID() == fac.Leader {
		replyErr(p, o, "§cYou cannot kick the faction leader.")
		return
	}
	if err := c.sessionManager.RemovePlayerFromFaction(target.UUID(), fac); err != nil {
		replyErr(p, o, "§cUnable to kick that member right now.")
		return
	}
	reply(p, o, "§aRemoved %s from the faction.", target.Name())
	target.Messagef("§cYou were removed from faction %s.", fac.Name)
}

func (c FactionPromote) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	fac, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		replyErr(p, o, "§cYou are not in a faction.")
		return
	}
	if fac.Leader != p.UUID() {
		replyErr(p, o, "§cOnly the faction leader can promote members.")
		return
	}
	if len(c.Target) != 1 {
		replyErr(p, o, "§cChoose exactly one online player.")
		return
	}
	target, ok := c.Target[0].(*player.Player)
	if !ok {
		replyErr(p, o, "§cThat target is invalid.")
		return
	}
	if _, member := fac.Members[target.UUID()]; !member {
		replyErr(p, o, "§cThat player is not in your faction.")
		return
	}
	if target.UUID() == fac.Leader {
		replyErr(p, o, "§cThe leader is already the highest rank.")
		return
	}
	if _, already := fac.Coleaders[target.UUID()]; already {
		replyErr(p, o, "§eThat player is already a co-leader.")
		return
	}
	fac.Coleaders[target.UUID()] = target.Name()
	if err := c.sessionManager.SaveFaction(fac); err != nil {
		replyErr(p, o, "§cUnable to promote that player right now.")
		return
	}
	reply(p, o, "§aPromoted %s to co-leader.", target.Name())
	target.Messagef("§aYou were promoted to co-leader in %s.", fac.Name)
}

func (c FactionDemote) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	fac, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		replyErr(p, o, "§cYou are not in a faction.")
		return
	}
	if fac.Leader != p.UUID() {
		replyErr(p, o, "§cOnly the faction leader can demote co-leaders.")
		return
	}
	if len(c.Target) != 1 {
		replyErr(p, o, "§cChoose exactly one online player.")
		return
	}
	target, ok := c.Target[0].(*player.Player)
	if !ok {
		replyErr(p, o, "§cThat target is invalid.")
		return
	}
	if _, already := fac.Coleaders[target.UUID()]; !already {
		replyErr(p, o, "§cThat player is not a co-leader.")
		return
	}
	delete(fac.Coleaders, target.UUID())
	if err := c.sessionManager.SaveFaction(fac); err != nil {
		replyErr(p, o, "§cUnable to demote that player right now.")
		return
	}
	reply(p, o, "§aDemoted %s to member.", target.Name())
	target.Messagef("§eYou were demoted to member in %s.", fac.Name)
}
