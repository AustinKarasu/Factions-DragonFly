package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/jorgebyte/faction/internal/faction"
	"github.com/jorgebyte/faction/internal/session"
	"time"
)

// AllyAction is an Enum for alliance actions.
type AllyAction string

func (AllyAction) Type() string { return "AllyAction" }
func (AllyAction) Options(src cmd.Source) []string {
	return []string{"send", "accept", "deny"}
}

// FactionAlly defines the structure for the /f ally <action> <faction> commands.
type FactionAlly struct {
	sessionManager *session.Manager
	Ally           cmd.SubCommand `cmd:"ally"`
	Action         AllyAction     `cmd:"action"`
	TargetFaction  string         `cmd:"faction"`
}

// Run is executed when a player uses /f ally.
func (c FactionAlly) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	p, ok := src.(*player.Player)
	if !ok {
		return
	}
	switch c.Action {
	case "send":
		c.handleSend(p, o)
	case "accept":
		c.handleAccept(p, o)
	case "deny":
		c.handleDeny(p, o)
	}
}

func (c FactionAlly) handleSend(p *player.Player, o *cmd.Output) {
	// Validations
	senderFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		o.Errorf("§cYou are not in a faction.")
		return
	}
	if senderFaction.Leader != p.UUID() {
		o.Errorf("§cOnly the leader can send alliance requests.")
		return
	}
	if len(senderFaction.Allies) >= faction.MaxAllies {
		o.Errorf("§cYour faction has already reached the alliance limit (%d).", faction.MaxAllies)
		return
	}
	receiverFaction, ok := c.sessionManager.GetFactionByName(c.TargetFaction)
	if !ok {
		o.Errorf("§cThe faction '%s' does not exist.", c.TargetFaction)
		return
	}
	if senderFaction.ID == receiverFaction.ID {
		o.Errorf("§cYou cannot ally with yourself.")
		return
	}
	if len(receiverFaction.Allies) >= faction.MaxAllies {
		o.Errorf("§cThe faction '%s' has already reached its alliance limit.", receiverFaction.Name)
		return
	}

	// Action
	c.sessionManager.AddAllianceRequest(senderFaction.ID, receiverFaction.ID, 5*time.Minute)
	o.Printf("§aAlliance request sent to '%s'. They have 5 minutes to accept.", receiverFaction.Name)
}

func (c FactionAlly) handleAccept(p *player.Player, o *cmd.Output) {
	receiverFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		o.Errorf("§cYou are not in a faction.")
		return
	}
	if receiverFaction.Leader != p.UUID() {
		o.Errorf("§cOnly the leader can accept alliances.")
		return
	}
	req, ok := c.sessionManager.GetAllianceRequest(receiverFaction.ID)
	if !ok {
		o.Errorf("§cYour faction has no pending alliance requests.")
		return
	}
	senderFaction, ok := c.sessionManager.GetFactionByID(req.SenderID)
	if !ok || senderFaction.Name != c.TargetFaction {
		o.Errorf("§cYou do not have an alliance request from '%s'.", c.TargetFaction)
		return
	}

	// Check limits again at the moment of acceptance
	if len(receiverFaction.Allies) >= faction.MaxAllies {
		o.Errorf("§cYour faction has reached the alliance limit (%d) and cannot accept more.", faction.MaxAllies)
		return
	}
	if len(senderFaction.Allies) >= faction.MaxAllies {
		o.Errorf("§cThe faction '%s' reached its alliance limit while you were waiting and cannot form a new one.", senderFaction.Name)
		return
	}

	// Action
	if err := c.sessionManager.FormAlliance(senderFaction, receiverFaction); err != nil {
		o.Errorf("§cAn error occurred while forming the alliance.")
		return
	}
	c.sessionManager.RemoveAllianceRequest(receiverFaction.ID)
	o.Printf("§aYour faction is now allied with '%s'!", senderFaction.Name)
}

func (c FactionAlly) handleDeny(p *player.Player, o *cmd.Output) {
	// Validations (similar to 'accept')
	receiverFaction, ok := c.sessionManager.GetPlayerFaction(p.UUID())
	if !ok {
		o.Errorf("§cYou are not in a faction.")
		return
	}
	if receiverFaction.Leader != p.UUID() {
		o.Errorf("§cOnly the leader can deny alliance requests.")
		return
	}
	req, ok := c.sessionManager.GetAllianceRequest(receiverFaction.ID)
	if !ok {
		o.Errorf("§cYour faction has no pending alliance requests.")
		return
	}
	senderFaction, ok := c.sessionManager.GetFactionByID(req.SenderID)
	if !ok || senderFaction.Name != c.TargetFaction {
		o.Errorf("§cYou do not have an alliance request from '%s'.", c.TargetFaction)
		return
	}

	// Action
	c.sessionManager.RemoveAllianceRequest(receiverFaction.ID)
	o.Printf("§eYou have denied the alliance request from '%s'.", senderFaction.Name)
}
