package npc

// Slot is a ranked NPC podium location.
type Slot struct {
	ID        int64
	Slot      int
	WorldName string
	X         float64
	Y         float64
	Z         float64
	Yaw       float64
	Pitch     float64
}
