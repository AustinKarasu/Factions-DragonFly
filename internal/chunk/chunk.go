package chunk

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
)

// Pos represents the position of a chunk in the world using X and Z coordinates.
type Pos struct {
	X, Z int32
}

// String returns a human-readable representation of the chunk position.
func (p Pos) String() string {
	return fmt.Sprintf("Chunk<%d, %d>", p.X, p.Z)
}

// FromWorldPos converts a world position (e.g., player position) to its corresponding chunk position.
func FromWorldPos(pos mgl64.Vec3) Pos {
	return Pos{
		X: int32(pos.X()) >> 4, // A chunk is 16x16 blocks, so divide by 16.
		Z: int32(pos.Z()) >> 4,
	}
}
