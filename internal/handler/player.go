package handler

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/jorgebyte/faction/internal/session"
)

// PlayerHandler handles player-specific events and delegates state updates.
type PlayerHandler struct {
	player.NopHandler

	p *player.Player
	s *session.Manager
}

// NewPlayerHandler creates a new PlayerHandler instance for the given player and session manager.
func NewPlayerHandler(p *player.Player, s *session.Manager) *PlayerHandler {
	return &PlayerHandler{p: p, s: s}
}

// HandleMove is triggered whenever the player moves.
func (h *PlayerHandler) HandleMove(ctx *player.Context, newPos mgl64.Vec3, newRot cube.Rotation) {
	h.s.UpdatePlayerChunk(h.p)
}

// HandleSkinChange keeps the NPC skin cache fresh.
func (h *PlayerHandler) HandleSkinChange(ctx *player.Context, newSkin *skin.Skin) {
	h.s.UpdateSkinCache(h.p.UUID(), *newSkin)
}

// HandleDeath updates faction/economy combat stats.
func (h *PlayerHandler) HandleDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
	h.s.HandlePlayerDeath(p, src)
}

// HandleItemUseOnEntity opens faction info when a ranked NPC is clicked.
func (h *PlayerHandler) HandleItemUseOnEntity(ctx *player.Context, e world.Entity) {
	if fac, ok := h.s.NPCFaction(e.H().UUID()); ok {
		h.p.Messagef("§6Top faction: §f%s §7Leader: §f%s", fac.Name, fac.Members[fac.Leader])
		h.p.ExecuteCommand("f info " + fac.Name)
	}
}

// HandleAttackEntity prevents players from damaging ranked NPCs.
func (h *PlayerHandler) HandleAttackEntity(ctx *player.Context, e world.Entity, force, height *float64, critical *bool) {
	if _, ok := h.s.NPCFaction(e.H().UUID()); ok {
		ctx.Cancel()
		h.p.Message("§eThese faction NPCs are protected.")
	}
}

// HandleQuit is triggered when the player disconnects from the server.
func (h *PlayerHandler) HandleQuit(p *player.Player) {
	h.s.HandlePlayerQuit(p)
}
