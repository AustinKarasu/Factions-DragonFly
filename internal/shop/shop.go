package shop

import (
	"strings"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Category is a player-facing grouping of shop items.
type Category struct {
	ID   int64
	Name string
	Icon string
	Sort int
}

// Entry is a configurable buy/sell listing in the shop.
type Entry struct {
	ID          int64
	CategoryID  int64
	Identifier  string
	Meta        int16
	Count       int
	BuyPrice    int
	SellPrice   int
	DisplayName string
	Lore        string
	Sort        int
}

// LoreLines splits the lore string into individual lines.
func (e Entry) LoreLines() []string {
	if strings.TrimSpace(e.Lore) == "" {
		return nil
	}
	return strings.Split(e.Lore, "\n")
}

// Stack attempts to build a Dragonfly item stack from the entry definition.
func (e Entry) Stack() (item.Stack, bool) {
	it, ok := world.ItemByName(e.Identifier, e.Meta)
	if !ok {
		return item.Stack{}, false
	}
	stack := item.NewStack(it, e.Count)
	if e.DisplayName != "" {
		stack = stack.WithCustomName(e.DisplayName)
	}
	if lore := e.LoreLines(); len(lore) > 0 {
		stack = stack.WithLore(lore...)
	}
	return stack, true
}

// Display returns the best shop label available for the entry.
func (e Entry) Display() string {
	if e.DisplayName != "" {
		return e.DisplayName
	}
	return e.Identifier
}
