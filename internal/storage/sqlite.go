package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jorgebyte/faction/internal/chunk"
	"github.com/jorgebyte/faction/internal/faction"
	"github.com/jorgebyte/faction/internal/npc"
	playerdata "github.com/jorgebyte/faction/internal/player"
	"github.com/jorgebyte/faction/internal/shop"
	_ "modernc.org/sqlite"
)

// Store defines the interface for persistent storage operations.
type Store interface {
	Init() error
	Close() error
	SaveFaction(f *faction.Faction) error
	GetAllFactions() ([]*faction.Faction, error)
	DeleteFaction(factionID uuid.UUID) error
	SavePlayer(p *playerdata.Data) error
	GetAllPlayers() ([]*playerdata.Data, error)
	GetAllClaims() (map[chunk.Pos]uuid.UUID, error)
	SaveClaim(pos chunk.Pos, factionID uuid.UUID) error
	DeleteClaim(pos chunk.Pos) error
	SaveShopCategory(category *shop.Category) error
	DeleteShopCategory(categoryID int64) error
	GetShopCategories() ([]shop.Category, error)
	SaveShopItem(entry *shop.Entry) error
	DeleteShopItem(itemID int64) error
	GetShopItems() ([]shop.Entry, error)
	GetNPCSlots() ([]npc.Slot, error)
}

// SQLiteStore provides an implementation of Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore initializes a new SQLiteStore using the given database path.
func NewSQLiteStore(path string) (Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

// Init creates the required tables if they do not already exist.
func (s *SQLiteStore) Init() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS factions (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			leader TEXT NOT NULL,
			power INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			coleaders TEXT,
			members TEXT,
			allies TEXT,
			claims INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS players (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			power INTEGER NOT NULL,
			balance INTEGER NOT NULL DEFAULT 500,
			kills INTEGER NOT NULL DEFAULT 0,
			deaths INTEGER NOT NULL DEFAULT 0,
			current_streak INTEGER NOT NULL DEFAULT 0,
			best_streak INTEGER NOT NULL DEFAULT 0,
			bounty INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS claims (
			chunk_x INTEGER,
			chunk_z INTEGER,
			faction_id TEXT NOT NULL,
			PRIMARY KEY(chunk_x, chunk_z)
		);`,
		`CREATE TABLE IF NOT EXISTS shop_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			icon TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS shop_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category_id INTEGER NOT NULL,
			identifier TEXT NOT NULL,
			meta INTEGER NOT NULL DEFAULT 0,
			count INTEGER NOT NULL DEFAULT 1,
			buy_price INTEGER NOT NULL DEFAULT 0,
			sell_price INTEGER NOT NULL DEFAULT 0,
			display_name TEXT NOT NULL DEFAULT '',
			lore TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY(category_id) REFERENCES shop_categories(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS npc_slots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slot_index INTEGER NOT NULL UNIQUE,
			world_name TEXT NOT NULL,
			x REAL NOT NULL,
			y REAL NOT NULL,
			z REAL NOT NULL,
			yaw REAL NOT NULL DEFAULT 0,
			pitch REAL NOT NULL DEFAULT 0
		);`,
		`ALTER TABLE factions ADD COLUMN description TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE players ADD COLUMN balance INTEGER NOT NULL DEFAULT 500;`,
		`ALTER TABLE players ADD COLUMN kills INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE players ADD COLUMN deaths INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE players ADD COLUMN current_streak INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE players ADD COLUMN best_streak INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE players ADD COLUMN bounty INTEGER NOT NULL DEFAULT 0;`,
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			if isDuplicateColumn(err) {
				continue
			}
			return fmt.Errorf("init statement failed: %w", err)
		}
	}
	if err := s.seedShop(); err != nil {
		return err
	}
	if err := s.seedNPCSlots(); err != nil {
		return err
	}
	return nil
}

func isDuplicateColumn(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate column name") || strings.Contains(err.Error(), "already exists"))
}

func (s *SQLiteStore) seedShop() error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM shop_categories`).Scan(&count); err != nil {
		return fmt.Errorf("count shop categories: %w", err)
	}
	if count > 0 {
		return nil
	}

	type seedCategory struct {
		name  string
		icon  string
		sort  int
		items []shop.Entry
	}
	seeds := []seedCategory{
		{
			name: "Resources", icon: "textures/items/diamond", sort: 1,
			items: []shop.Entry{
				{Identifier: "minecraft:diamond", Count: 1, BuyPrice: 250, SellPrice: 175, DisplayName: "Diamond", Sort: 1},
				{Identifier: "minecraft:gold_ingot", Count: 1, BuyPrice: 60, SellPrice: 40, DisplayName: "Gold Ingot", Sort: 2},
				{Identifier: "minecraft:iron_ingot", Count: 1, BuyPrice: 45, SellPrice: 30, DisplayName: "Iron Ingot", Sort: 3},
			},
		},
		{
			name: "Food", icon: "textures/items/bread", sort: 2,
			items: []shop.Entry{
				{Identifier: "minecraft:bread", Count: 8, BuyPrice: 55, SellPrice: 25, DisplayName: "Bread Bundle", Sort: 1},
				{Identifier: "minecraft:cooked_beef", Count: 8, BuyPrice: 95, SellPrice: 45, DisplayName: "Steak Bundle", Sort: 2},
				{Identifier: "minecraft:golden_carrot", Count: 8, BuyPrice: 140, SellPrice: 60, DisplayName: "Golden Carrot Bundle", Sort: 3},
			},
		},
		{
			name: "Combat", icon: "textures/items/iron_sword", sort: 3,
			items: []shop.Entry{
				{Identifier: "minecraft:iron_sword", Count: 1, BuyPrice: 175, SellPrice: 90, DisplayName: "Iron Sword", Sort: 1},
				{Identifier: "minecraft:bow", Count: 1, BuyPrice: 140, SellPrice: 70, DisplayName: "Bow", Sort: 2},
				{Identifier: "minecraft:arrow", Count: 16, BuyPrice: 50, SellPrice: 20, DisplayName: "Arrow Pack", Sort: 3},
			},
		},
	}

	for _, category := range seeds {
		res, err := s.db.Exec(`INSERT INTO shop_categories (name, icon, sort_order) VALUES (?, ?, ?)`, category.name, category.icon, category.sort)
		if err != nil {
			return fmt.Errorf("insert seed category %s: %w", category.name, err)
		}
		categoryID, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("last category id: %w", err)
		}
		for _, entry := range category.items {
			if _, err = s.db.Exec(
				`INSERT INTO shop_items (category_id, identifier, meta, count, buy_price, sell_price, display_name, lore, sort_order)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				categoryID, entry.Identifier, entry.Meta, entry.Count, entry.BuyPrice, entry.SellPrice, entry.DisplayName, entry.Lore, entry.Sort,
			); err != nil {
				return fmt.Errorf("insert seed item %s: %w", entry.Identifier, err)
			}
		}
	}
	return nil
}

func (s *SQLiteStore) seedNPCSlots() error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM npc_slots`).Scan(&count); err != nil {
		return fmt.Errorf("count npc slots: %w", err)
	}
	if count > 0 {
		return nil
	}

	slots := []npc.Slot{
		{Slot: 1, WorldName: "World", X: 0, Y: 65, Z: 6, Yaw: 180},
		{Slot: 2, WorldName: "World", X: -3, Y: 65, Z: 8, Yaw: 180},
		{Slot: 3, WorldName: "World", X: 3, Y: 65, Z: 8, Yaw: 180},
		{Slot: 4, WorldName: "World", X: -6, Y: 65, Z: 10, Yaw: 180},
		{Slot: 5, WorldName: "World", X: 6, Y: 65, Z: 10, Yaw: 180},
	}
	for _, slot := range slots {
		if _, err := s.db.Exec(
			`INSERT INTO npc_slots (slot_index, world_name, x, y, z, yaw, pitch) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			slot.Slot, slot.WorldName, slot.X, slot.Y, slot.Z, slot.Yaw, slot.Pitch,
		); err != nil {
			return fmt.Errorf("insert npc slot %d: %w", slot.Slot, err)
		}
	}
	return nil
}

// Close terminates the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// SaveFaction inserts or updates a faction record in the database.
func (s *SQLiteStore) SaveFaction(f *faction.Faction) error {
	coleadersJSON, _ := json.Marshal(f.Coleaders)
	membersJSON, _ := json.Marshal(f.Members)
	alliesJSON, _ := json.Marshal(f.Allies)

	query := `
	INSERT INTO factions (id, name, leader, power, created_at, description, coleaders, members, allies, claims)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		leader = excluded.leader,
		power = excluded.power,
		description = excluded.description,
		coleaders = excluded.coleaders,
		members = excluded.members,
		allies = excluded.allies,
		claims = excluded.claims;`

	_, err := s.db.Exec(query,
		f.ID.String(),
		f.Name,
		f.Leader.String(),
		f.Power,
		f.CreatedAt.Format(time.RFC3339),
		f.Description,
		string(coleadersJSON),
		string(membersJSON),
		string(alliesJSON),
		f.Claims,
	)
	return err
}

// GetAllFactions retrieves all faction records from the database.
func (s *SQLiteStore) GetAllFactions() ([]*faction.Faction, error) {
	rows, err := s.db.Query(`
		SELECT id, name, leader, power, created_at, description, coleaders, members, allies, claims
		FROM factions`)
	if err != nil {
		return nil, fmt.Errorf("failed to query factions: %w", err)
	}
	defer rows.Close()

	var factions []*faction.Faction
	for rows.Next() {
		var f faction.Faction
		var id, leader, createdAt string
		var coleadersJSON, membersJSON, alliesJSON []byte

		if err := rows.Scan(&id, &f.Name, &leader, &f.Power, &createdAt, &f.Description, &coleadersJSON, &membersJSON, &alliesJSON, &f.Claims); err != nil {
			return nil, fmt.Errorf("scan faction row: %w", err)
		}

		f.ID, _ = uuid.Parse(id)
		f.Leader, _ = uuid.Parse(leader)
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		_ = json.Unmarshal(coleadersJSON, &f.Coleaders)
		_ = json.Unmarshal(membersJSON, &f.Members)
		_ = json.Unmarshal(alliesJSON, &f.Allies)
		if f.Coleaders == nil {
			f.Coleaders = make(map[uuid.UUID]string)
		}
		if f.Members == nil {
			f.Members = make(map[uuid.UUID]string)
		}
		if f.Allies == nil {
			f.Allies = make(map[uuid.UUID]string)
		}
		factions = append(factions, &f)
	}
	return factions, nil
}

// DeleteFaction removes a faction from the database using its ID.
func (s *SQLiteStore) DeleteFaction(factionID uuid.UUID) error {
	_, err := s.db.Exec(`DELETE FROM factions WHERE id = ?`, factionID.String())
	return err
}

// SavePlayer inserts or updates a player's data in the database.
func (s *SQLiteStore) SavePlayer(p *playerdata.Data) error {
	query := `
	INSERT INTO players (uuid, name, power, balance, kills, deaths, current_streak, best_streak, bounty)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(uuid) DO UPDATE SET
		name = excluded.name,
		power = excluded.power,
		balance = excluded.balance,
		kills = excluded.kills,
		deaths = excluded.deaths,
		current_streak = excluded.current_streak,
		best_streak = excluded.best_streak,
		bounty = excluded.bounty;`
	_, err := s.db.Exec(query, p.UUID.String(), p.Name, p.Power, p.Balance, p.Kills, p.Deaths, p.CurrentStreak, p.BestStreak, p.Bounty)
	return err
}

// GetAllPlayers retrieves all player data from the database.
func (s *SQLiteStore) GetAllPlayers() ([]*playerdata.Data, error) {
	rows, err := s.db.Query(`
		SELECT uuid, name, power, balance, kills, deaths, current_streak, best_streak, bounty
		FROM players`)
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	var players []*playerdata.Data
	for rows.Next() {
		var p playerdata.Data
		var uuidStr string
		if err := rows.Scan(&uuidStr, &p.Name, &p.Power, &p.Balance, &p.Kills, &p.Deaths, &p.CurrentStreak, &p.BestStreak, &p.Bounty); err != nil {
			return nil, fmt.Errorf("scan player row: %w", err)
		}
		p.UUID, _ = uuid.Parse(uuidStr)
		players = append(players, &p)
	}
	return players, nil
}

// GetAllClaims retrieves all chunk claims and their owning factions.
func (s *SQLiteStore) GetAllClaims() (map[chunk.Pos]uuid.UUID, error) {
	rows, err := s.db.Query(`SELECT chunk_x, chunk_z, faction_id FROM claims`)
	if err != nil {
		return nil, fmt.Errorf("failed to query claims: %w", err)
	}
	defer rows.Close()

	claims := make(map[chunk.Pos]uuid.UUID)
	for rows.Next() {
		var pos chunk.Pos
		var factionID string
		if err := rows.Scan(&pos.X, &pos.Z, &factionID); err != nil {
			return nil, fmt.Errorf("scan claim row: %w", err)
		}
		id, _ := uuid.Parse(factionID)
		claims[pos] = id
	}
	return claims, nil
}

// SaveClaim inserts a new claim into the database.
func (s *SQLiteStore) SaveClaim(pos chunk.Pos, factionID uuid.UUID) error {
	_, err := s.db.Exec(`INSERT INTO claims (chunk_x, chunk_z, faction_id) VALUES (?, ?, ?)`, pos.X, pos.Z, factionID.String())
	return err
}

// DeleteClaim removes a specific chunk claim from the database.
func (s *SQLiteStore) DeleteClaim(pos chunk.Pos) error {
	_, err := s.db.Exec(`DELETE FROM claims WHERE chunk_x = ? AND chunk_z = ?`, pos.X, pos.Z)
	return err
}

// SaveShopCategory inserts or updates a shop category.
func (s *SQLiteStore) SaveShopCategory(category *shop.Category) error {
	if category.ID == 0 {
		res, err := s.db.Exec(`INSERT INTO shop_categories (name, icon, sort_order) VALUES (?, ?, ?)`, category.Name, category.Icon, category.Sort)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		category.ID = id
		return nil
	}
	_, err := s.db.Exec(`UPDATE shop_categories SET name = ?, icon = ?, sort_order = ? WHERE id = ?`, category.Name, category.Icon, category.Sort, category.ID)
	return err
}

// DeleteShopCategory removes a category and its items.
func (s *SQLiteStore) DeleteShopCategory(categoryID int64) error {
	if _, err := s.db.Exec(`DELETE FROM shop_items WHERE category_id = ?`, categoryID); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM shop_categories WHERE id = ?`, categoryID)
	return err
}

// GetShopCategories returns all configured shop categories.
func (s *SQLiteStore) GetShopCategories() ([]shop.Category, error) {
	rows, err := s.db.Query(`SELECT id, name, icon, sort_order FROM shop_categories ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []shop.Category
	for rows.Next() {
		var category shop.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Icon, &category.Sort); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

// SaveShopItem inserts or updates a shop item.
func (s *SQLiteStore) SaveShopItem(entry *shop.Entry) error {
	if entry.ID == 0 {
		res, err := s.db.Exec(
			`INSERT INTO shop_items (category_id, identifier, meta, count, buy_price, sell_price, display_name, lore, sort_order)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			entry.CategoryID, entry.Identifier, entry.Meta, entry.Count, entry.BuyPrice, entry.SellPrice, entry.DisplayName, entry.Lore, entry.Sort,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		entry.ID = id
		return nil
	}
	_, err := s.db.Exec(
		`UPDATE shop_items
		 SET category_id = ?, identifier = ?, meta = ?, count = ?, buy_price = ?, sell_price = ?, display_name = ?, lore = ?, sort_order = ?
		 WHERE id = ?`,
		entry.CategoryID, entry.Identifier, entry.Meta, entry.Count, entry.BuyPrice, entry.SellPrice, entry.DisplayName, entry.Lore, entry.Sort, entry.ID,
	)
	return err
}

// DeleteShopItem removes a shop item.
func (s *SQLiteStore) DeleteShopItem(itemID int64) error {
	_, err := s.db.Exec(`DELETE FROM shop_items WHERE id = ?`, itemID)
	return err
}

// GetShopItems returns all configured shop items.
func (s *SQLiteStore) GetShopItems() ([]shop.Entry, error) {
	rows, err := s.db.Query(`
		SELECT id, category_id, identifier, meta, count, buy_price, sell_price, display_name, lore, sort_order
		FROM shop_items
		ORDER BY category_id, sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []shop.Entry
	for rows.Next() {
		var entry shop.Entry
		if err := rows.Scan(&entry.ID, &entry.CategoryID, &entry.Identifier, &entry.Meta, &entry.Count, &entry.BuyPrice, &entry.SellPrice, &entry.DisplayName, &entry.Lore, &entry.Sort); err != nil {
			return nil, err
		}
		items = append(items, entry)
	}
	return items, nil
}

// GetNPCSlots returns all configured NPC slots.
func (s *SQLiteStore) GetNPCSlots() ([]npc.Slot, error) {
	rows, err := s.db.Query(`SELECT id, slot_index, world_name, x, y, z, yaw, pitch FROM npc_slots ORDER BY slot_index`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []npc.Slot
	for rows.Next() {
		var slot npc.Slot
		if err := rows.Scan(&slot.ID, &slot.Slot, &slot.WorldName, &slot.X, &slot.Y, &slot.Z, &slot.Yaw, &slot.Pitch); err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}
	return slots, nil
}
