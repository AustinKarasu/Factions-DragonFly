package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/jorgebyte/faction/internal/command"
	"github.com/jorgebyte/faction/internal/handler"
	"github.com/jorgebyte/faction/internal/session"
	"github.com/jorgebyte/faction/internal/storage"
	tasks "github.com/jorgebyte/faction/internal/task"
	"github.com/pelletier/go-toml"
)

func main() {
	// ---SETUP LOGGER ---
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	chat.Global.Subscribe(chat.StdoutSubscriber{})

	// --- INITIALIZE STORAGE ---
	store, err := storage.NewSQLiteStore("factions.db")
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		return
	}
	defer store.Close()

	if err := store.Init(); err != nil {
		log.Error("Failed to initialize database tables", "error", err)
		return
	}
	log.Info("Database connected and initialized.")

	// --- INITIALIZE SESSION MANAGER (IN-MEMORY CACHE) ---
	sessionManager := session.NewManager(store)

	if err := sessionManager.LoadAllFactions(); err != nil {
		log.Error("Failed to load factions into memory", "error", err)
		return
	}
	if err := sessionManager.LoadAllPlayers(); err != nil {
		log.Error("Failed to load player data into memory", "error", err)
		return
	}
	if err := sessionManager.LoadAllClaims(); err != nil {
		log.Error("Failed to load claims into memory", "error", err)
		return
	}
	if err := sessionManager.LoadShopData(); err != nil {
		log.Error("Failed to load shop data", "error", err)
		return
	}
	if err := sessionManager.LoadNPCSlots(); err != nil {
		log.Error("Failed to load NPC slots", "error", err)
		return
	}
	log.Info("Faction and player caches loaded.")

	// --- REGISTER COMMANDS ---
	command.RegisterAll(sessionManager)
	log.Info("Commands registered.")

	// --- LOAD CONFIGURATION & START SERVER ---
	conf, err := readConfig(log)
	if err != nil {
		panic(fmt.Errorf("critical error reading config: %w", err))
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()

	log.Info("Factions starting...")
	srv.Listen()
	// Periodically show borders to viewers.
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for range ticker.C {
			for p := range srv.Players(nil) {
				if sessionManager.IsBorderViewer(p.UUID()) {
					tasks.ShowChunkBorders(p, sessionManager)
				}
			}
		}
	}()

	scoreboardTicker := time.NewTicker(3 * time.Second)
	go func() {
		for range scoreboardTicker.C {
			sessionManager.UpdateAllScoreboards()
		}
	}()

	// Handle new player joins.
	for p := range srv.Accept() {
		p.Handle(handler.NewPlayerHandler(p, sessionManager))
		sessionManager.GetOrCreatePlayer(p.UUID(), p.Name())
		sessionManager.RegisterOnlinePlayer(p)
	}
}

// readConfig reads the configuration from the config.toml file, or creates the
// file if it does not yet exist.
func readConfig(log *slog.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return zero, fmt.Errorf("create default config: %v", err)
		}
		return c.Config(log)
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return zero, fmt.Errorf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return zero, fmt.Errorf("decode config: %v", err)
	}
	return c.Config(log)
}
