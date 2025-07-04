package player

import "github.com/google/uuid"

// Data represents persistent player data such as UUID, name, and power level.
type Data struct {
	UUID  uuid.UUID
	Name  string
	Power int
}

// New creates a new player data instance with default values.
func New(uuid uuid.UUID, name string) *Data {
	return &Data{
		UUID:  uuid,
		Name:  name,
		Power: 2, // Default starting power.
	}
}
