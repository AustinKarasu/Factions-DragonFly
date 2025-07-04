package faction

import (
	"time"

	"github.com/google/uuid"
	"github.com/jorgebyte/faction/internal/player"
)

// MaxAllies defines the maximum number of allied factions allowed.
const MaxAllies = 1

// MaxClaims defines the maximum number of territory claims a faction can hold.
const MaxClaims = 1

// Faction represents a group of players with shared power, claims, and relationships.
type Faction struct {
	ID        uuid.UUID            `json:"id"`
	Name      string               `json:"name"`
	Leader    uuid.UUID            `json:"leader"`
	Power     int                  `json:"power"`
	CreatedAt time.Time            `json:"created_at"`
	Coleaders map[uuid.UUID]string `json:"coleaders"` // UUID → Player Name
	Members   map[uuid.UUID]string `json:"members"`   // UUID → Player Name
	Claims    int                  `json:"claims"`    // Total number of claimed chunks
	Allies    map[uuid.UUID]string `json:"allies"`    // UUID → Faction Name
}

// New creates and initializes a new faction instance with the given leader.
func New(name string, leaderID uuid.UUID, leaderName string) *Faction {
	return &Faction{
		ID:        uuid.New(),
		Name:      name,
		Leader:    leaderID,
		Power:     2,
		CreatedAt: time.Now(),
		Coleaders: make(map[uuid.UUID]string),
		Members:   map[uuid.UUID]string{leaderID: leaderName},
		Claims:    0,
		Allies:    make(map[uuid.UUID]string),
	}
}

// CalculatePower calculates the total power of the faction, including all its members.
func (f *Faction) CalculatePower(allPlayers map[uuid.UUID]*player.Data) int {
	totalPower := f.Power
	for memberID := range f.Members {
		if data, ok := allPlayers[memberID]; ok {
			totalPower += data.Power
		}
	}
	return totalPower
}
