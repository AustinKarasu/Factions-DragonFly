package faction

import (
	"time"

	"github.com/google/uuid"
	playerdata "github.com/jorgebyte/faction/internal/player"
)

// MaxAllies defines the maximum number of allied factions allowed.
const MaxAllies = 3

// Faction represents a faction with members, allies, and territorial power.
type Faction struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Leader      uuid.UUID            `json:"leader"`
	Power       int                  `json:"power"`
	CreatedAt   time.Time            `json:"created_at"`
	Description string               `json:"description"`
	Coleaders   map[uuid.UUID]string `json:"coleaders"`
	Members     map[uuid.UUID]string `json:"members"`
	Claims      int                  `json:"claims"`
	Allies      map[uuid.UUID]string `json:"allies"`
}

// New creates and initializes a new faction instance with the given leader.
func New(name string, leaderID uuid.UUID, leaderName string) *Faction {
	return &Faction{
		ID:          uuid.New(),
		Name:        name,
		Leader:      leaderID,
		Power:       0,
		CreatedAt:   time.Now(),
		Description: "Newly forged faction.",
		Coleaders:   make(map[uuid.UUID]string),
		Members:     map[uuid.UUID]string{leaderID: leaderName},
		Claims:      0,
		Allies:      make(map[uuid.UUID]string),
	}
}

// CalculatePower calculates the total power of the faction, including all its members.
func (f *Faction) CalculatePower(allPlayers map[uuid.UUID]*playerdata.Data) int {
	totalPower := f.Power
	for memberID := range f.Members {
		if data, ok := allPlayers[memberID]; ok {
			totalPower += data.Power
		}
	}
	return totalPower
}

// MemberCount returns the total number of faction members.
func (f *Faction) MemberCount() int {
	return len(f.Members)
}

// ClaimLimit returns the total amount of claims the faction may hold.
func (f *Faction) ClaimLimit() int {
	limit := f.MemberCount() * 4
	if limit < 8 {
		limit = 8
	}
	return limit
}
