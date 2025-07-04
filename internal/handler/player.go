package handler

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/jorgebyte/faction/internal/session"
)

// PlayerHandler handles player-specific events and delegates state updates.
type PlayerHandler struct {
	player.NopHandler // Embeds NopHandler to inherit default behavior.

	p *player.Player
	s *session.Manager
}

// NewPlayerHandler creates a new PlayerHandler instance for the given player and session manager.
func NewPlayerHandler(p *player.Player, s *session.Manager) *PlayerHandler {
	return &PlayerHandler{p: p, s: s}
}

// HandleMove is triggered whenever the player moves.
// It updates the chunk state and triggers any chunk-based logic.
func (h *PlayerHandler) HandleMove(ctx *player.Context, newPos mgl64.Vec3, newRot cube.Rotation) {
	h.s.UpdatePlayerChunk(h.p)
}

// HandleQuit is triggered when the player disconnects from the server.
// It performs cleanup such as removing state from memory.
func (h *PlayerHandler) HandleQuit(p *player.Player) {
	h.s.HandlePlayerQuit(p)
}
