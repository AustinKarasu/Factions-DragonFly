package tasks

import (
	"image/color"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/jorgebyte/faction/internal/chunk"
	"github.com/jorgebyte/faction/internal/session"
)

// Particle colors used for visualizing chunk borders.
var (
	colorClaimed = color.RGBA{R: 255, A: 255} // Red for claimed chunks.
	colorFree    = color.RGBA{G: 255, A: 255} // Green for unclaimed chunks.
)

// ShowChunkBorders scans the surrounding chunks around the player and displays their borders.
// Claimed chunks are shown in red, unclaimed in green.
func ShowChunkBorders(p *player.Player, s *session.Manager) {
	playerChunk := chunk.FromWorldPos(p.Position())
	viewRadius := int32(2) // Display a 5x5 chunk area (2 chunks in all directions).

	for x := playerChunk.X - viewRadius; x <= playerChunk.X+viewRadius; x++ {
		for z := playerChunk.Z - viewRadius; z <= playerChunk.Z+viewRadius; z++ {
			currentPos := chunk.Pos{X: x, Z: z}
			var borderColor color.RGBA

			if _, claimed := s.GetClaimOwner(currentPos); claimed {
				borderColor = colorClaimed
			} else {
				borderColor = colorFree
			}

			drawChunkBorder(p, currentPos, p.Position().Y()+0.4, borderColor)
		}
	}
}

// drawChunkBorder renders particles along the edges of a single chunk at a specified height.
func drawChunkBorder(p *player.Player, pos chunk.Pos, y float64, c color.RGBA) {
	startX := float64(pos.X * 16)
	endX := startX + 16.0
	startZ := float64(pos.Z * 16)
	endZ := startZ + 16.0

	dust := particle.Dust{Colour: c}

	// Render the four edges of the chunk using particles.
	for i := 0.0; i < 16.0; i += 1.0 {
		// North edge (Z fixed)
		p.ShowParticle(mgl64.Vec3{startX + i, y, startZ}, dust)
		// South edge
		p.ShowParticle(mgl64.Vec3{startX + i, y, endZ}, dust)
		// West edge (X fixed)
		p.ShowParticle(mgl64.Vec3{startX, y, startZ + i}, dust)
		// East edge
		p.ShowParticle(mgl64.Vec3{endX, y, startZ + i}, dust)
	}
}
