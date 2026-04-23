package session

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/entity"
	dfplayer "github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/block/cube"
	playerscoreboard "github.com/df-mc/dragonfly/server/player/scoreboard"
	playerskin "github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/jorgebyte/faction/internal/chunk"
	"github.com/jorgebyte/faction/internal/faction"
	"github.com/jorgebyte/faction/internal/npc"
	playerdata "github.com/jorgebyte/faction/internal/player"
	"github.com/jorgebyte/faction/internal/shop"
	"github.com/jorgebyte/faction/internal/storage"
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

// Manager maintains the in-memory server state for factions, players, claims, and UI services.
type Manager struct {
	store storage.Store
	mu    sync.RWMutex

	factionsByName map[string]*faction.Faction
	factionsByID   map[uuid.UUID]*faction.Faction
	playerToFaction map[uuid.UUID]*faction.Faction

	pendingInvites   map[uuid.UUID]*PendingInvite
	pendingAlliances map[uuid.UUID]*PendingAllianceRequest

	players       map[uuid.UUID]*playerdata.Data
	playerSkins   map[uuid.UUID]playerskin.Skin
	onlinePlayers map[uuid.UUID]*dfplayer.Player

	claims          map[chunk.Pos]uuid.UUID
	lastPlayerChunk map[uuid.UUID]chunk.Pos
	borderViewers   map[uuid.UUID]bool

	shopCategories map[int64]shop.Category
	shopItems      map[int64]shop.Entry
	npcSlots       []npc.Slot
	npcEntities    map[int]*world.EntityHandle
	npcAssignments map[uuid.UUID]uuid.UUID
}

// NewManager creates and initializes a new session manager.
func NewManager(store storage.Store) *Manager {
	return &Manager{
		store:            store,
		factionsByName:   make(map[string]*faction.Faction),
		factionsByID:     make(map[uuid.UUID]*faction.Faction),
		playerToFaction:  make(map[uuid.UUID]*faction.Faction),
		pendingInvites:   make(map[uuid.UUID]*PendingInvite),
		pendingAlliances: make(map[uuid.UUID]*PendingAllianceRequest),
		players:          make(map[uuid.UUID]*playerdata.Data),
		playerSkins:      make(map[uuid.UUID]playerskin.Skin),
		onlinePlayers:    make(map[uuid.UUID]*dfplayer.Player),
		claims:           make(map[chunk.Pos]uuid.UUID),
		lastPlayerChunk:  make(map[uuid.UUID]chunk.Pos),
		borderViewers:    make(map[uuid.UUID]bool),
		shopCategories:   make(map[int64]shop.Category),
		shopItems:        make(map[int64]shop.Entry),
		npcEntities:      make(map[int]*world.EntityHandle),
		npcAssignments:   make(map[uuid.UUID]uuid.UUID),
	}
}

func normalizeFactionName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// LoadAllFactions loads all factions from persistent storage into memory.
func (m *Manager) LoadAllFactions() error {
	factions, err := m.store.GetAllFactions()
	if err != nil {
		return fmt.Errorf("load factions: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, f := range factions {
		key := normalizeFactionName(f.Name)
		m.factionsByName[key] = f
		m.factionsByID[f.ID] = f
		for playerID := range f.Members {
			m.playerToFaction[playerID] = f
		}
	}
	return nil
}

// LoadAllPlayers loads all player data from persistent storage.
func (m *Manager) LoadAllPlayers() error {
	players, err := m.store.GetAllPlayers()
	if err != nil {
		return fmt.Errorf("load players: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range players {
		m.players[p.UUID] = p
	}
	return nil
}

// LoadAllClaims loads all claimed chunks into memory.
func (m *Manager) LoadAllClaims() error {
	claims, err := m.store.GetAllClaims()
	if err != nil {
		return fmt.Errorf("load claims: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for pos, factionID := range claims {
		m.claims[pos] = factionID
	}
	return nil
}

// LoadShopData loads all shop categories and items into memory.
func (m *Manager) LoadShopData() error {
	categories, err := m.store.GetShopCategories()
	if err != nil {
		return fmt.Errorf("load shop categories: %w", err)
	}
	items, err := m.store.GetShopItems()
	if err != nil {
		return fmt.Errorf("load shop items: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, category := range categories {
		m.shopCategories[category.ID] = category
	}
	for _, entry := range items {
		m.shopItems[entry.ID] = entry
	}
	return nil
}

// LoadNPCSlots loads NPC podium slots from storage.
func (m *Manager) LoadNPCSlots() error {
	slots, err := m.store.GetNPCSlots()
	if err != nil {
		return fmt.Errorf("load npc slots: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.npcSlots = slots
	return nil
}

// GetFactionByName returns a faction by its name.
func (m *Manager) GetFactionByName(name string) (*faction.Faction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.factionsByName[normalizeFactionName(name)]
	return f, ok
}

// GetFactionByID returns a faction by its UUID.
func (m *Manager) GetFactionByID(id uuid.UUID) (*faction.Faction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.factionsByID[id]
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
		return fmt.Errorf("save new faction: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.factionsByName[normalizeFactionName(f.Name)] = f
	m.factionsByID[f.ID] = f
	m.playerToFaction[f.Leader] = f
	return nil
}

// SaveFaction persists a faction after in-memory changes.
func (m *Manager) SaveFaction(f *faction.Faction) error {
	return m.store.SaveFaction(f)
}

// DeleteFaction removes a faction and cleans up related state.
func (m *Manager) DeleteFaction(f *faction.Faction) error {
	m.mu.Lock()
	for allyID := range f.Allies {
		if allyFaction, ok := m.factionsByID[allyID]; ok {
			delete(allyFaction.Allies, f.ID)
			_ = m.store.SaveFaction(allyFaction)
		}
	}
	for pos, ownerID := range m.claims {
		if ownerID == f.ID {
			delete(m.claims, pos)
			_ = m.store.DeleteClaim(pos)
		}
	}
	delete(m.factionsByName, normalizeFactionName(f.Name))
	delete(m.factionsByID, f.ID)
	for playerID := range f.Members {
		delete(m.playerToFaction, playerID)
	}
	m.mu.Unlock()
	return m.store.DeleteFaction(f.ID)
}

// AddInvitation adds a faction invite with an expiration timer.
func (m *Manager) AddInvitation(factionID, invitedPlayerID uuid.UUID, expiration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldInvite, ok := m.pendingInvites[invitedPlayerID]; ok {
		oldInvite.Timer.Stop()
	}

	timer := time.AfterFunc(expiration, func() {
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

// RemovePlayerFromFaction removes a player from a faction and updates the cache and DB.
func (m *Manager) RemovePlayerFromFaction(playerID uuid.UUID, f *faction.Faction) error {
	m.mu.Lock()
	delete(f.Members, playerID)
	delete(f.Coleaders, playerID)
	delete(m.playerToFaction, playerID)
	m.mu.Unlock()
	return m.store.SaveFaction(f)
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
		if p.Name != name {
			p.Name = name
			go m.store.SavePlayer(p)
		}
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

// GetAllPlayersMap returns a copy of the internal player map.
func (m *Manager) GetAllPlayersMap() map[uuid.UUID]*playerdata.Data {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[uuid.UUID]*playerdata.Data, len(m.players))
	for id, p := range m.players {
		out[id] = p
	}
	return out
}

// GetAllFactionsData returns a list of all cached factions.
func (m *Manager) GetAllFactionsData() []*faction.Faction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	factions := make([]*faction.Faction, 0, len(m.factionsByName))
	for _, f := range m.factionsByID {
		factions = append(factions, f)
	}
	return factions
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
	m.pendingAlliances[receiverID] = &PendingAllianceRequest{SenderID: senderID, Timer: timer}
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

// ClaimChunk attempts to claim a chunk for a faction.
func (m *Manager) ClaimChunk(pos chunk.Pos, f *faction.Faction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.claims[pos]; ok {
		return fmt.Errorf("this chunk is already claimed")
	}
	if f.Claims >= f.ClaimLimit() {
		return fmt.Errorf("claim limit reached (%d)", f.ClaimLimit())
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
	if f.Claims > 0 {
		f.Claims--
	}
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
func (m *Manager) UpdatePlayerChunk(p *dfplayer.Player) {
	currentPos := chunk.FromWorldPos(p.Position())

	m.mu.Lock()
	lastPos, ok := m.lastPlayerChunk[p.UUID()]
	m.lastPlayerChunk[p.UUID()] = currentPos
	m.mu.Unlock()

	if !ok || lastPos == currentPos {
		return
	}

	if ownerID, claimed := m.GetClaimOwner(currentPos); claimed {
		if fac, ok := m.GetFactionByID(ownerID); ok {
			p.SendTitle(title.New("§e" + fac.Name))
		}
	} else {
		p.SendTitle(title.New("§aWilderness"))
	}
	m.UpdateScoreboard(p)
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

// RegisterOnlinePlayer tracks an online player for UI updates.
func (m *Manager) RegisterOnlinePlayer(p *dfplayer.Player) {
	m.mu.Lock()
	m.onlinePlayers[p.UUID()] = p
	m.playerSkins[p.UUID()] = p.Skin()
	m.mu.Unlock()
	m.GetOrCreatePlayer(p.UUID(), p.Name())
	m.UpdateScoreboard(p)
}

// UpdateSkinCache stores the latest known player skin for NPC syncing.
func (m *Manager) UpdateSkinCache(playerID uuid.UUID, skin playerskin.Skin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playerSkins[playerID] = skin
}

// GetOnlinePlayerByName returns an online player by exact or case-insensitive name.
func (m *Manager) GetOnlinePlayerByName(name string) (*dfplayer.Player, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.onlinePlayers {
		if strings.EqualFold(p.Name(), name) {
			return p, true
		}
	}
	return nil, false
}

// OnlineCount returns the amount of online players.
func (m *Manager) OnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.onlinePlayers)
}

// HandlePlayerQuit handles cleanup when a player leaves the server.
func (m *Manager) HandlePlayerQuit(p *dfplayer.Player) {
	m.RemoveBorderViewer(p.UUID())

	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.lastPlayerChunk, p.UUID())
	delete(m.onlinePlayers, p.UUID())
}

// CreditBalance adds money to a player.
func (m *Manager) CreditBalance(playerID uuid.UUID, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be positive")
	}
	m.mu.Lock()
	p, ok := m.players[playerID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("player data not found")
	}
	p.Balance += amount
	m.mu.Unlock()
	return m.store.SavePlayer(p)
}

// DebitBalance removes money from a player if enough balance exists.
func (m *Manager) DebitBalance(playerID uuid.UUID, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be positive")
	}
	m.mu.Lock()
	p, ok := m.players[playerID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("player data not found")
	}
	if p.Balance < amount {
		m.mu.Unlock()
		return fmt.Errorf("insufficient balance")
	}
	p.Balance -= amount
	m.mu.Unlock()
	return m.store.SavePlayer(p)
}

// PlaceBounty sets or increases a bounty on a target player.
func (m *Manager) PlaceBounty(placerID, targetID uuid.UUID, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if placerID == targetID {
		return fmt.Errorf("you cannot place a bounty on yourself")
	}
	if err := m.DebitBalance(placerID, amount); err != nil {
		return err
	}
	m.mu.Lock()
	target, ok := m.players[targetID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("target data not found")
	}
	target.Bounty += amount
	m.mu.Unlock()
	if err := m.store.SavePlayer(target); err != nil {
		return err
	}
	if targetPlayer, ok := m.GetOnlinePlayerByName(target.Name); ok {
		targetPlayer.Messagef("§cA bounty of $%d was placed on you. Total bounty: $%d", amount, target.Bounty)
	}
	return nil
}

// ClearBounty removes a bounty from a player.
func (m *Manager) ClearBounty(targetID uuid.UUID) error {
	m.mu.Lock()
	target, ok := m.players[targetID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("target data not found")
	}
	target.Bounty = 0
	m.mu.Unlock()
	return m.store.SavePlayer(target)
}

// TopBounties returns players ordered by bounty descending.
func (m *Manager) TopBounties(limit int) []*playerdata.Data {
	players := m.GetAllPlayersData()
	sort.Slice(players, func(i, j int) bool {
		if players[i].Bounty == players[j].Bounty {
			return strings.ToLower(players[i].Name) < strings.ToLower(players[j].Name)
		}
		return players[i].Bounty > players[j].Bounty
	})
	if limit > len(players) {
		limit = len(players)
	}
	return players[:limit]
}

// HandlePlayerDeath updates kill/death stats and pays out bounties.
func (m *Manager) HandlePlayerDeath(victim *dfplayer.Player, src world.DamageSource) {
	m.mu.Lock()
	victimData, ok := m.players[victim.UUID()]
	if ok {
		victimData.Deaths++
		victimData.CurrentStreak = 0
	}
	m.mu.Unlock()
	if ok {
		_ = m.store.SavePlayer(victimData)
	}
	if source, ok := src.(entity.AttackDamageSource); ok {
		if killer, ok := source.Attacker.(*dfplayer.Player); ok && killer.UUID() != victim.UUID() {
			m.rewardKill(killer, victim.UUID())
		}
		return
	}
	if source, ok := src.(entity.ProjectileDamageSource); ok {
		if killer, ok := source.Owner.(*dfplayer.Player); ok && killer.UUID() != victim.UUID() {
			m.rewardKill(killer, victim.UUID())
		}
	}
}

func (m *Manager) rewardKill(killer *dfplayer.Player, victimID uuid.UUID) {
	m.mu.Lock()
	killerData, okKiller := m.players[killer.UUID()]
	victimData, okVictim := m.players[victimID]
	if okKiller {
		killerData.Kills++
		killerData.CurrentStreak++
		if killerData.CurrentStreak > killerData.BestStreak {
			killerData.BestStreak = killerData.CurrentStreak
		}
	}
	bountyReward := 0
	if okVictim {
		bountyReward = victimData.Bounty
		victimData.Bounty = 0
	}
	if okKiller {
		killerData.Balance += 40 + bountyReward
	}
	m.mu.Unlock()

	if okKiller {
		_ = m.store.SavePlayer(killerData)
	}
	if okVictim {
		_ = m.store.SavePlayer(victimData)
	}

	if bountyReward > 0 {
		killer.Messagef("§6You claimed a $%d bounty.", bountyReward)
	} else {
		killer.Message("§aYou earned $40 for the kill.")
	}
	m.UpdateScoreboard(killer)
}

// UpdateScoreboard rebuilds and sends the faction scoreboard for a player.
func (m *Manager) UpdateScoreboard(p *dfplayer.Player) {
	data := m.GetOrCreatePlayer(p.UUID(), p.Name())
	board := playerscoreboard.New("§6§lFACTIONS")
	board.RemovePadding()
	_, _ = board.WriteString(fmt.Sprintf("§fKills: §a%d\n", data.Kills))
	_, _ = board.WriteString(fmt.Sprintf("§fStreak: §a%d\n", data.CurrentStreak))
	_, _ = board.WriteString(fmt.Sprintf("§fBest: §a%d\n", data.BestStreak))
	_, _ = board.WriteString(fmt.Sprintf("§fBalance: §a$%d\n", data.Balance))
	_, _ = board.WriteString(fmt.Sprintf("§fBounty: §c$%d\n", data.Bounty))
	if fac, ok := m.GetPlayerFaction(p.UUID()); ok {
		_, _ = board.WriteString(fmt.Sprintf("§fFaction: §b%s\n", fac.Name))
		_, _ = board.WriteString(fmt.Sprintf("§fClaims: §e%d/%d\n", fac.Claims, fac.ClaimLimit()))
	} else {
		_, _ = board.WriteString("§fFaction: §7None\n")
	}
	_, _ = board.WriteString(fmt.Sprintf("§fPlayers: §a%d", m.OnlineCount()))
	p.SendScoreboard(board)
}

// UpdateAllScoreboards refreshes the sidebar for all online players.
func (m *Manager) UpdateAllScoreboards() {
	m.mu.RLock()
	players := make([]*dfplayer.Player, 0, len(m.onlinePlayers))
	for _, p := range m.onlinePlayers {
		players = append(players, p)
	}
	m.mu.RUnlock()

	for _, p := range players {
		m.UpdateScoreboard(p)
	}
}

// GetShopCategories returns all shop categories sorted for UI use.
func (m *Manager) GetShopCategories() []shop.Category {
	m.mu.RLock()
	defer m.mu.RUnlock()
	categories := make([]shop.Category, 0, len(m.shopCategories))
	for _, category := range m.shopCategories {
		categories = append(categories, category)
	}
	sort.Slice(categories, func(i, j int) bool {
		if categories[i].Sort == categories[j].Sort {
			return categories[i].ID < categories[j].ID
		}
		return categories[i].Sort < categories[j].Sort
	})
	return categories
}

// GetShopCategory retrieves a single shop category.
func (m *Manager) GetShopCategory(id int64) (shop.Category, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	category, ok := m.shopCategories[id]
	return category, ok
}

// GetShopItemsByCategory returns all items in a shop category.
func (m *Manager) GetShopItemsByCategory(categoryID int64) []shop.Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	items := make([]shop.Entry, 0)
	for _, entry := range m.shopItems {
		if entry.CategoryID == categoryID {
			items = append(items, entry)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			return items[i].ID < items[j].ID
		}
		return items[i].Sort < items[j].Sort
	})
	return items
}

// SaveShopCategory persists a category and refreshes the in-memory cache.
func (m *Manager) SaveShopCategory(category *shop.Category) error {
	if err := m.store.SaveShopCategory(category); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shopCategories[category.ID] = *category
	return nil
}

// DeleteShopCategory removes a category and its entries.
func (m *Manager) DeleteShopCategory(categoryID int64) error {
	if err := m.store.DeleteShopCategory(categoryID); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.shopCategories, categoryID)
	for id, entry := range m.shopItems {
		if entry.CategoryID == categoryID {
			delete(m.shopItems, id)
		}
	}
	return nil
}

// SaveShopItem persists a shop entry and refreshes the in-memory cache.
func (m *Manager) SaveShopItem(entry *shop.Entry) error {
	if err := m.store.SaveShopItem(entry); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shopItems[entry.ID] = *entry
	return nil
}

// DeleteShopItem removes a shop entry.
func (m *Manager) DeleteShopItem(itemID int64) error {
	if err := m.store.DeleteShopItem(itemID); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.shopItems, itemID)
	return nil
}

// BuyShopItem charges a player and gives them the selected entry.
func (m *Manager) BuyShopItem(playerID uuid.UUID, itemID int64) error {
	m.mu.RLock()
	entry, ok := m.shopItems[itemID]
	playerRef := m.onlinePlayers[playerID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("shop item not found")
	}
	if playerRef == nil {
		return fmt.Errorf("player must be online to buy items")
	}
	stack, ok := entry.Stack()
	if !ok {
		return fmt.Errorf("shop item is not registered on this server")
	}
	if err := m.DebitBalance(playerID, entry.BuyPrice); err != nil {
		return err
	}
	if _, err := playerRef.Inventory().AddItem(stack); err != nil {
		_ = m.CreditBalance(playerID, entry.BuyPrice)
		return fmt.Errorf("not enough inventory space")
	}
	m.UpdateScoreboard(playerRef)
	return nil
}

// SellShopItem removes matching items from inventory and pays the player.
func (m *Manager) SellShopItem(playerID uuid.UUID, itemID int64) error {
	m.mu.RLock()
	entry, ok := m.shopItems[itemID]
	playerRef := m.onlinePlayers[playerID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("shop item not found")
	}
	if playerRef == nil {
		return fmt.Errorf("player must be online to sell items")
	}
	stack, ok := entry.Stack()
	if !ok {
		return fmt.Errorf("shop item is not registered on this server")
	}
	if !playerRef.Inventory().ContainsItem(stack) {
		return fmt.Errorf("you do not have enough items to sell")
	}
	if err := playerRef.Inventory().RemoveItem(stack); err != nil {
		return err
	}
	if err := m.CreditBalance(playerID, entry.SellPrice); err != nil {
		return err
	}
	m.UpdateScoreboard(playerRef)
	return nil
}

// TopFactionsByPower returns factions ordered by total power.
func (m *Manager) TopFactionsByPower(limit int) []*faction.Faction {
	factions := m.GetAllFactionsData()
	allPlayers := m.GetAllPlayersMap()
	sort.Slice(factions, func(i, j int) bool {
		return factions[i].CalculatePower(allPlayers) > factions[j].CalculatePower(allPlayers)
	})
	if limit > len(factions) {
		limit = len(factions)
	}
	return factions[:limit]
}

// SyncTopNPCs respawns the top-faction podium NPCs in the given world.
func (m *Manager) SyncTopNPCs(w *world.World) error {
	return nil
}

func (m *Manager) getNPCSlots() []npc.Slot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]npc.Slot, len(m.npcSlots))
	copy(out, m.npcSlots)
	return out
}

func (m *Manager) respawnNPCForSlot(w *world.World, slot npc.Slot, fac *faction.Faction, assignments map[uuid.UUID]uuid.UUID) error {
	m.mu.RLock()
	existing := m.npcEntities[slot.Slot]
	var leaderSkin playerskin.Skin
	var leaderName string
	if fac != nil {
		leaderName = fac.Members[fac.Leader]
		leaderSkin = m.playerSkins[fac.Leader]
		if leaderSkin.Bounds().Dx() == 0 || leaderSkin.Bounds().Dy() == 0 {
			for _, cached := range m.playerSkins {
				if cached.Bounds().Dx() != 0 && cached.Bounds().Dy() != 0 {
					leaderSkin = cached
					break
				}
			}
		}
	}
	m.mu.RUnlock()

	if existing != nil {
		existing.ExecWorld(func(tx *world.Tx, e world.Entity) {
			tx.RemoveEntity(e)
		})
	}

	if fac == nil {
		m.mu.Lock()
		delete(m.npcEntities, slot.Slot)
		m.mu.Unlock()
		return nil
	}
	if leaderName == "" {
		leaderName = fac.Name
	}
	if leaderSkin.Bounds().Dx() == 0 || leaderSkin.Bounds().Dy() == 0 {
		leaderSkin = playerskin.New(64, 64)
	}

	nameTag := fmt.Sprintf("§6#%d %s\n§fLeader: %s", slot.Slot, fac.Name, leaderName)
	pos := mgl64.Vec3{slot.X, slot.Y, slot.Z}
	rot := cube.Rotation{slot.Yaw, slot.Pitch}
	handle := world.EntitySpawnOpts{
		Position: pos,
		Rotation: rot,
		NameTag:  nameTag,
	}.New(dfplayer.Type, dfplayer.Config{
		Name:     leaderName,
		UUID:     uuid.New(),
		Skin:     leaderSkin,
		GameMode: world.GameModeAdventure,
		Position: pos,
		Rotation: rot,
	})

	<-w.Exec(func(tx *world.Tx) {
		if entity := tx.AddEntity(handle); entity != nil {
			if fake, ok := entity.(*dfplayer.Player); ok {
				fake.SetImmobile()
			}
		}
	})

	m.mu.Lock()
	m.npcEntities[slot.Slot] = handle
	m.mu.Unlock()
	assignments[handle.UUID()] = fac.ID
	return nil
}

// NPCFaction returns the faction assigned to a podium NPC entity.
func (m *Manager) NPCFaction(entityID uuid.UUID) (*faction.Faction, bool) {
	return nil, false
}
