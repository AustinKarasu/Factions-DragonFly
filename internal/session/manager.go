package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/google/uuid"
	"github.com/jorgebyte/faction/internal/chunk"
	"github.com/jorgebyte/faction/internal/faction"
	playerdata "github.com/jorgebyte/faction/internal/player"
	"github.com/jorgebyte/faction/internal/storage"
	"log/slog"
	"sync"
	"time"
)

// PendingInvite represents a pending faction invitation with an expiration timer.
type PendingInvite struct {
	FactionID uuid.UUID
	Timer     *time.Timer
}

// PendingAllianceRequest represents a pending alliance request between factions.
type PendingAllianceRequest struct {
	SenderID uuid.UUID
	Timer    *time.Timer
}

// Manager maintains the in-memory server state for factions, players, claims, and more.
type Manager struct {
	store storage.Store
	mu    sync.RWMutex

	factionsByName   map[string]*faction.Faction
	playerToFaction  map[uuid.UUID]*faction.Faction
	pendingInvites   map[uuid.UUID]*PendingInvite
	pendingAlliances map[uuid.UUID]*PendingAllianceRequest
	players          map[uuid.UUID]*playerdata.Data
	claims           map[chunk.Pos]uuid.UUID
	lastPlayerChunk  map[uuid.UUID]chunk.Pos
	borderViewers    map[uuid.UUID]bool
}

// NewManager creates and initializes a new session manager.
func NewManager(store storage.Store) *Manager {
	return &Manager{
		store:            store,
		factionsByName:   make(map[string]*faction.Faction),
		playerToFaction:  make(map[uuid.UUID]*faction.Faction),
		pendingInvites:   make(map[uuid.UUID]*PendingInvite),
		pendingAlliances: make(map[uuid.UUID]*PendingAllianceRequest),
		players:          make(map[uuid.UUID]*playerdata.Data),
		claims:           make(map[chunk.Pos]uuid.UUID),
		lastPlayerChunk:  make(map[uuid.UUID]chunk.Pos),
		borderViewers:    make(map[uuid.UUID]bool),
	}
}

// LoadAllFactions loads all factions from persistent storage into memory.
func (m *Manager) LoadAllFactions() error {
	factions, err := m.store.GetAllFactions()
	if err != nil {
		return fmt.Errorf("failed to load factions: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, f := range factions {
		m.factionsByName[f.Name] = f
		for playerID := range f.Members {
			m.playerToFaction[playerID] = f
		}
	}
	slog.Info("Factions loaded into memory", "count", len(factions))
	return nil
}

// GetFactionByName returns a faction by its name.
func (m *Manager) GetFactionByName(name string) (*faction.Faction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.factionsByName[name]
	return f, ok
}

// GetPlayerFaction returns the faction a player belongs to.
func (m *Manager) GetPlayerFaction(playerID uuid.UUID) (*faction.Faction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.playerToFaction[playerID]
	return f, ok
}

// CreateFaction persists and caches a newly created faction.
func (m *Manager) CreateFaction(f *faction.Faction) error {
	if err := m.store.SaveFaction(f); err != nil {
		return fmt.Errorf("failed to save new faction: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.factionsByName[f.Name] = f
	m.playerToFaction[f.Leader] = f
	return nil
}

// DeleteFaction removes a faction and cleans up related state.
func (m *Manager) DeleteFaction(f *faction.Faction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for allyID := range f.Allies {
		if allyFaction, ok := m.factionsByName[f.Allies[allyID]]; ok {
			delete(allyFaction.Allies, f.ID)
			go m.store.SaveFaction(allyFaction)
		}
	}

	if err := m.store.DeleteFaction(f.ID); err != nil {
		return fmt.Errorf("failed to delete faction from DB: %w", err)
	}

	delete(m.factionsByName, f.Name)
	for playerID := range f.Members {
		delete(m.playerToFaction, playerID)
	}
	return nil
}

// AddInvitation adds a faction invite with an expiration timer.
func (m *Manager) AddInvitation(factionID, invitedPlayerID uuid.UUID, expiration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldInvite, ok := m.pendingInvites[invitedPlayerID]; ok {
		oldInvite.Timer.Stop()
	}

	timer := time.AfterFunc(expiration, func() {
		slog.Info("Invitation expired", "playerID", invitedPlayerID)
		m.RemoveInvitation(invitedPlayerID)
	})

	m.pendingInvites[invitedPlayerID] = &PendingInvite{
		FactionID: factionID,
		Timer:     timer,
	}
}

// GetInvitation returns the pending invitation for a player.
func (m *Manager) GetInvitation(invitedPlayerID uuid.UUID) (*PendingInvite, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	invite, ok := m.pendingInvites[invitedPlayerID]
	return invite, ok
}

// RemoveInvitation cancels and deletes a pending invitation.
func (m *Manager) RemoveInvitation(invitedPlayerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if invite, ok := m.pendingInvites[invitedPlayerID]; ok {
		invite.Timer.Stop()
		delete(m.pendingInvites, invitedPlayerID)
	}
}

// AddPlayerToFaction adds a player to a faction and persists the change.
func (m *Manager) AddPlayerToFaction(playerID uuid.UUID, playerName string, f *faction.Faction) error {
	m.mu.Lock()
	f.Members[playerID] = playerName
	m.playerToFaction[playerID] = f
	m.mu.Unlock()

	return m.store.SaveFaction(f)
}

// GetFactionByID returns a faction by its UUID (inefficient search).
func (m *Manager) GetFactionByID(id uuid.UUID) (*faction.Faction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, f := range m.factionsByName {
		if f.ID == id {
			return f, true
		}
	}
	return nil, false
}

// LoadAllPlayers loads all player data from persistent storage.
func (m *Manager) LoadAllPlayers() error {
	players, err := m.store.GetAllPlayers()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range players {
		m.players[p.UUID] = p
	}
	slog.Info("Player data loaded into memory", "count", len(players))
	return nil
}

// GetPlayer returns cached player data.
func (m *Manager) GetPlayer(playerID uuid.UUID) (*playerdata.Data, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.players[playerID]
	return p, ok
}

// GetOrCreatePlayer returns a player or creates a new entry if not found.
func (m *Manager) GetOrCreatePlayer(id uuid.UUID, name string) *playerdata.Data {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.players[id]; ok {
		return p
	}
	p := playerdata.New(id, name)
	m.players[id] = p
	go m.store.SavePlayer(p)
	return p
}

// GetAllPlayersData returns a slice of all cached player data.
func (m *Manager) GetAllPlayersData() []*playerdata.Data {
	m.mu.RLock()
	defer m.mu.RUnlock()
	players := make([]*playerdata.Data, 0, len(m.players))
	for _, p := range m.players {
		players = append(players, p)
	}
	return players
}

// RemovePlayerFromFaction removes a player from a faction and updates the cache and DB.
func (m *Manager) RemovePlayerFromFaction(playerID uuid.UUID, f *faction.Faction) error {
	m.mu.Lock()
	delete(f.Members, playerID)
	delete(f.Coleaders, playerID)
	delete(m.playerToFaction, playerID)
	m.mu.Unlock()

	return m.store.SaveFaction(f)
}

// GetAllFactionsData returns a list of all cached factions.
func (m *Manager) GetAllFactionsData() []*faction.Faction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	factions := make([]*faction.Faction, 0, len(m.factionsByName))
	for _, f := range m.factionsByName {
		factions = append(factions, f)
	}
	return factions
}

// GetAllPlayersMap returns a copy of the internal player map.
func (m *Manager) GetAllPlayersMap() map[uuid.UUID]*playerdata.Data {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copy := make(map[uuid.UUID]*playerdata.Data, len(m.players))
	for id, p := range m.players {
		copy[id] = p
	}
	return copy
}

// AddAllianceRequest creates a timed alliance request between two factions.
func (m *Manager) AddAllianceRequest(senderID, receiverID uuid.UUID, expiration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if old, ok := m.pendingAlliances[receiverID]; ok {
		old.Timer.Stop()
	}

	timer := time.AfterFunc(expiration, func() {
		m.RemoveAllianceRequest(receiverID)
	})

	m.pendingAlliances[receiverID] = &PendingAllianceRequest{
		SenderID: senderID,
		Timer:    timer,
	}
}

// GetAllianceRequest returns a pending alliance request.
func (m *Manager) GetAllianceRequest(receiverID uuid.UUID) (*PendingAllianceRequest, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	req, ok := m.pendingAlliances[receiverID]
	return req, ok
}

// RemoveAllianceRequest deletes an existing alliance request.
func (m *Manager) RemoveAllianceRequest(receiverID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if req, ok := m.pendingAlliances[receiverID]; ok {
		req.Timer.Stop()
		delete(m.pendingAlliances, receiverID)
	}
}

// FormAlliance creates a mutual alliance between two factions.
func (m *Manager) FormAlliance(facA, facB *faction.Faction) error {
	m.mu.Lock()
	facA.Allies[facB.ID] = facB.Name
	facB.Allies[facA.ID] = facA.Name
	m.mu.Unlock()

	if err := m.store.SaveFaction(facA); err != nil {
		return err
	}
	return m.store.SaveFaction(facB)
}

// LoadAllClaims loads all claimed chunks into memory.
func (m *Manager) LoadAllClaims() error {
	claims, err := m.store.GetAllClaims()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for pos, factionID := range claims {
		m.claims[pos] = factionID
	}
	slog.Info("Claims loaded into memory", "count", len(m.claims))
	return nil
}

// ClaimChunk attempts to claim a chunk for a faction.
func (m *Manager) ClaimChunk(pos chunk.Pos, f *faction.Faction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.claims[pos]; ok {
		return fmt.Errorf("this chunk is already claimed")
	}
	if f.Claims >= faction.MaxClaims {
		return fmt.Errorf("claim limit reached (%d)", faction.MaxClaims)
	}
	f.Claims++
	m.claims[pos] = f.ID
	go m.store.SaveClaim(pos, f.ID)
	go m.store.SaveFaction(f)
	return nil
}

// UnclaimChunk removes a claim from a chunk.
func (m *Manager) UnclaimChunk(pos chunk.Pos, f *faction.Faction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ownerID, ok := m.claims[pos]
	if !ok {
		return fmt.Errorf("this chunk is not claimed")
	}
	if ownerID != f.ID {
		return fmt.Errorf("your faction does not own this chunk")
	}
	f.Claims--
	delete(m.claims, pos)
	go m.store.DeleteClaim(pos)
	go m.store.SaveFaction(f)
	return nil
}

// GetClaimOwner returns the faction ID that owns a chunk, if any.
func (m *Manager) GetClaimOwner(pos chunk.Pos) (uuid.UUID, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.claims[pos]
	return id, ok
}

// UpdatePlayerChunk tracks player movement and sends title updates.
func (m *Manager) UpdatePlayerChunk(p *player.Player) {
	currentPos := chunk.FromWorldPos(p.Position())

	m.mu.Lock()
	lastPos, ok := m.lastPlayerChunk[p.UUID()]
	m.lastPlayerChunk[p.UUID()] = currentPos
	m.mu.Unlock()

	if !ok || lastPos == currentPos {
		return
	}

	if ownerID, ok := m.GetClaimOwner(currentPos); ok {
		if faction, ok := m.GetFactionByID(ownerID); ok {
			p.SendTitle(title.New(fmt.Sprintf("§e%s", faction.Name)))
			return
		}
	}
	p.SendTitle(title.New("§aWilderness"))
}

// ToggleBorderView toggles chunk border visualization for a player.
func (m *Manager) ToggleBorderView(playerID uuid.UUID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	newState := !m.borderViewers[playerID]
	m.borderViewers[playerID] = newState
	return newState
}

// IsBorderViewer checks if the player has border view enabled.
func (m *Manager) IsBorderViewer(playerID uuid.UUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.borderViewers[playerID]
}

// RemoveBorderViewer cleans up border view state when a player disconnects.
func (m *Manager) RemoveBorderViewer(playerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.borderViewers, playerID)
}

// HandlePlayerQuit handles cleanup when a player leaves the server.
func (m *Manager) HandlePlayerQuit(p *player.Player) {
	m.RemoveBorderViewer(p.UUID())

	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.lastPlayerChunk, p.UUID())
}
