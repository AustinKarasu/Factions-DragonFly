package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jorgebyte/faction/internal/chunk"
	"github.com/jorgebyte/faction/internal/faction"
	"github.com/jorgebyte/faction/internal/player"
	_ "modernc.org/sqlite"
)

// Store defines the interface for persistent storage operations.
type Store interface {
	Init() error
	Close() error
	SaveFaction(f *faction.Faction) error
	GetAllFactions() ([]*faction.Faction, error)
	DeleteFaction(factionID uuid.UUID) error
	SavePlayer(p *player.Data) error
	GetAllPlayers() ([]*player.Data, error)
	GetAllClaims() (map[chunk.Pos]uuid.UUID, error)
	SaveClaim(pos chunk.Pos, factionID uuid.UUID) error
	DeleteClaim(pos chunk.Pos) error
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
	factionQuery := `
    CREATE TABLE IF NOT EXISTS factions (
       id TEXT PRIMARY KEY,
       name TEXT UNIQUE NOT NULL,
       leader TEXT NOT NULL,
       power INTEGER NOT NULL,
       created_at TEXT NOT NULL,
       coleaders TEXT,
       members TEXT,
       allies TEXT,
       claims INTEGER NOT NULL DEFAULT 0
    );`
	if _, err := s.db.Exec(factionQuery); err != nil {
		return fmt.Errorf("create factions table: %w", err)
	}

	playerQuery := `
    CREATE TABLE IF NOT EXISTS players (
       uuid TEXT PRIMARY KEY,
       name TEXT NOT NULL,
       power INTEGER NOT NULL
    );`
	if _, err := s.db.Exec(playerQuery); err != nil {
		return fmt.Errorf("create players table: %w", err)
	}

	claimQuery := `
    CREATE TABLE IF NOT EXISTS claims (
       chunk_x INTEGER,
       chunk_z INTEGER,
       faction_id TEXT NOT NULL,
       PRIMARY KEY(chunk_x, chunk_z)
    );`
	_, err := s.db.Exec(claimQuery)
	return err
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
    INSERT INTO factions (id, name, leader, power, created_at, coleaders, members, allies, claims)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) 
    ON CONFLICT(id) DO UPDATE SET
       name = excluded.name,
       leader = excluded.leader,
       power = excluded.power,
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
		SELECT id, name, leader, power, created_at, coleaders, members, allies, claims
		FROM factions`)
	if err != nil {
		return nil, fmt.Errorf("failed to query factions: %w", err)
	}
	defer rows.Close()

	var factions []*faction.Faction
	for rows.Next() {
		var f faction.Faction
		var id, leader, createdAtStr string
		var coleadersJSON, membersJSON, alliesJSON []byte

		if err := rows.Scan(
			&id,
			&f.Name,
			&leader,
			&f.Power,
			&createdAtStr,
			&coleadersJSON,
			&membersJSON,
			&alliesJSON,
			&f.Claims,
		); err != nil {
			return nil, fmt.Errorf("failed to scan faction row: %w", err)
		}

		f.ID, _ = uuid.Parse(id)
		f.Leader, _ = uuid.Parse(leader)
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		_ = json.Unmarshal(coleadersJSON, &f.Coleaders)
		_ = json.Unmarshal(membersJSON, &f.Members)
		if err := json.Unmarshal(alliesJSON, &f.Allies); err != nil {
			f.Allies = make(map[uuid.UUID]string)
		}

		factions = append(factions, &f)
	}
	return factions, nil
}

// DeleteFaction removes a faction from the database using its ID.
func (s *SQLiteStore) DeleteFaction(factionID uuid.UUID) error {
	query := `DELETE FROM factions WHERE id = ?`
	_, err := s.db.Exec(query, factionID.String())
	return err
}

// SavePlayer inserts or updates a player's data in the database.
func (s *SQLiteStore) SavePlayer(p *player.Data) error {
	query := `
	INSERT INTO players (uuid, name, power)
	VALUES (?, ?, ?)
	ON CONFLICT(uuid) DO UPDATE SET
		name = excluded.name,
		power = excluded.power;`
	_, err := s.db.Exec(query, p.UUID.String(), p.Name, p.Power)
	return err
}

// GetAllPlayers retrieves all player data from the database.
func (s *SQLiteStore) GetAllPlayers() ([]*player.Data, error) {
	rows, err := s.db.Query("SELECT uuid, name, power FROM players")
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	var players []*player.Data
	for rows.Next() {
		var p player.Data
		var uuidStr string
		if err := rows.Scan(&uuidStr, &p.Name, &p.Power); err != nil {
			return nil, fmt.Errorf("failed to scan player row: %w", err)
		}
		p.UUID, _ = uuid.Parse(uuidStr)
		players = append(players, &p)
	}
	return players, nil
}

// GetAllClaims retrieves all chunk claims and their owning factions.
func (s *SQLiteStore) GetAllClaims() (map[chunk.Pos]uuid.UUID, error) {
	rows, err := s.db.Query("SELECT chunk_x, chunk_z, faction_id FROM claims")
	if err != nil {
		return nil, fmt.Errorf("failed to query claims: %w", err)
	}
	defer rows.Close()

	claims := make(map[chunk.Pos]uuid.UUID)
	for rows.Next() {
		var pos chunk.Pos
		var factionIDStr string
		if err := rows.Scan(&pos.X, &pos.Z, &factionIDStr); err != nil {
			return nil, fmt.Errorf("failed to scan claim row: %w", err)
		}
		factionID, _ := uuid.Parse(factionIDStr)
		claims[pos] = factionID
	}
	return claims, nil
}

// SaveClaim inserts a new claim into the database.
func (s *SQLiteStore) SaveClaim(pos chunk.Pos, factionID uuid.UUID) error {
	query := `INSERT INTO claims (chunk_x, chunk_z, faction_id) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, pos.X, pos.Z, factionID.String())
	return err
}

// DeleteClaim removes a specific chunk claim from the database.
func (s *SQLiteStore) DeleteClaim(pos chunk.Pos) error {
	query := `DELETE FROM claims WHERE chunk_x = ? AND chunk_z = ?`
	_, err := s.db.Exec(query, pos.X, pos.Z)
	return err
}
