package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/jorgebyte/faction/internal/faction"
	"github.com/jorgebyte/faction/internal/session"
	"github.com/jorgebyte/faction/internal/shop"
)

type menuSubmit struct {
	onSubmit func(submitter *player.Player, pressed form.Button, tx *world.Tx)
	onClose  func(submitter *player.Player, tx *world.Tx)
}

func (m menuSubmit) Submit(submitter form.Submitter, pressed form.Button, tx *world.Tx) {
	if p, ok := submitter.(*player.Player); ok && m.onSubmit != nil {
		m.onSubmit(p, pressed, tx)
	}
}

func (m menuSubmit) Close(submitter form.Submitter, tx *world.Tx) {
	if p, ok := submitter.(*player.Player); ok && m.onClose != nil {
		m.onClose(p, tx)
	}
}

type factionCreateForm struct {
	sessionManager *session.Manager
	player         *player.Player

	Name        form.Input
	Description form.Input
}

func newFactionCreateForm(p *player.Player, s *session.Manager) factionCreateForm {
	return factionCreateForm{
		sessionManager: s,
		player:         p,
		Name:           form.NewInput("Faction name", "", "Enter a faction name"),
		Description:    form.NewInput("Description", "", "What is your faction about?"),
	}
}

func (f factionCreateForm) Submit(submitter form.Submitter, tx *world.Tx) {
	name := strings.TrimSpace(f.Name.Value())
	if name == "" {
		f.player.Message("Ã‚Â§cFaction name cannot be empty.")
		return
	}
	if _, ok := f.sessionManager.GetPlayerFaction(f.player.UUID()); ok {
		f.player.Message("Ã‚Â§cYou are already in a faction.")
		return
	}
	if _, ok := f.sessionManager.GetFactionByName(name); ok {
		f.player.Messagef("Ã‚Â§cThe faction '%s' already exists.", name)
		return
	}
	newFaction := faction.New(name, f.player.UUID(), f.player.Name())
	if description := strings.TrimSpace(f.Description.Value()); description != "" {
		newFaction.Description = description
	}
	if err := f.sessionManager.CreateFaction(newFaction); err != nil {
		f.player.Message("Ã‚Â§cUnable to create faction right now.")
		return
	}
	f.player.Messagef("Ã‚Â§aFaction '%s' created successfully.", newFaction.Name)
	f.sessionManager.UpdateScoreboard(f.player)
}

type bountyPlaceForm struct {
	sessionManager *session.Manager
	player         *player.Player
	targets        []string
	targetIDs      []string

	Target form.Dropdown
	Amount form.Input
}

func newBountyPlaceForm(p *player.Player, s *session.Manager) bountyPlaceForm {
	targets := make([]string, 0)
	targetIDs := make([]string, 0)
	for _, data := range s.GetAllPlayersData() {
		if data.UUID == p.UUID() {
			continue
		}
		targets = append(targets, data.Name)
		targetIDs = append(targetIDs, data.UUID.String())
	}
	if len(targets) == 0 {
		targets = []string{"No targets available"}
		targetIDs = []string{""}
	}
	return bountyPlaceForm{
		sessionManager: s,
		player:         p,
		targets:        targets,
		targetIDs:      targetIDs,
		Target:         form.NewDropdown("Target", targets, 0),
		Amount:         form.NewInput("Amount", "", "Amount in dollars"),
	}
}

func (f bountyPlaceForm) Submit(submitter form.Submitter, tx *world.Tx) {
	if len(f.targetIDs) == 0 || f.targetIDs[f.Target.Value()] == "" {
		f.player.Message("Ã‚Â§cThere are no valid bounty targets.")
		return
	}
	targetID, err := parseUUID(f.targetIDs[f.Target.Value()])
	if err != nil {
		f.player.Message("Ã‚Â§cThat target is invalid.")
		return
	}
	amount, err := strconv.Atoi(strings.TrimSpace(f.Amount.Value()))
	if err != nil || amount <= 0 {
		f.player.Message("Ã‚Â§cEnter a valid bounty amount.")
		return
	}
	if err := f.sessionManager.PlaceBounty(f.player.UUID(), targetID, amount); err != nil {
		f.player.Messagef("Ã‚Â§cUnable to place bounty: %v", err)
		return
	}
	f.player.Messagef("Ã‚Â§aPlaced a $%d bounty on %s.", amount, f.targets[f.Target.Value()])
	f.sessionManager.UpdateScoreboard(f.player)
}

type shopCategoryForm struct {
	sessionManager *session.Manager
	player         *player.Player

	Name form.Input
	Icon form.Input
	Sort form.Input
}

func newShopCategoryForm(p *player.Player, s *session.Manager) shopCategoryForm {
	return shopCategoryForm{
		sessionManager: s,
		player:         p,
		Name:           form.NewInput("Category name", "", "Resources"),
		Icon:           form.NewInput("Button icon", "", "textures/items/diamond"),
		Sort:           form.NewInput("Sort order", "10", "10"),
	}
}

func (f shopCategoryForm) Submit(submitter form.Submitter, tx *world.Tx) {
	name := strings.TrimSpace(f.Name.Value())
	if name == "" {
		f.player.Message("Ã‚Â§cCategory name cannot be empty.")
		return
	}
	sortOrder, _ := strconv.Atoi(strings.TrimSpace(f.Sort.Value()))
	category := &shop.Category{
		Name: name,
		Icon: strings.TrimSpace(f.Icon.Value()),
		Sort: sortOrder,
	}
	if err := f.sessionManager.SaveShopCategory(category); err != nil {
		f.player.Messagef("Ã‚Â§cUnable to save category: %v", err)
		return
	}
	f.player.Messagef("Ã‚Â§aSaved shop category '%s'.", category.Name)
	openShopAdminMenu(f.player, f.sessionManager)
}

type shopItemForm struct {
	sessionManager *session.Manager
	player         *player.Player
	categoryIDs    []int64

	Category    form.Dropdown
	Identifier  form.Input
	Meta        form.Input
	Count       form.Input
	BuyPrice    form.Input
	SellPrice   form.Input
	DisplayName form.Input
	Lore        form.Input
	Sort        form.Input
}

type shopDeleteCategoryForm struct {
	sessionManager *session.Manager
	player         *player.Player
	categoryIDs    []int64

	Category form.Dropdown
}

func newShopDeleteCategoryForm(p *player.Player, s *session.Manager) shopDeleteCategoryForm {
	categories := s.GetShopCategories()
	names := make([]string, 0, len(categories))
	ids := make([]int64, 0, len(categories))
	for _, category := range categories {
		names = append(names, category.Name)
		ids = append(ids, category.ID)
	}
	if len(names) == 0 {
		names = []string{"No categories"}
		ids = []int64{0}
	}
	return shopDeleteCategoryForm{
		sessionManager: s,
		player:         p,
		categoryIDs:    ids,
		Category:       form.NewDropdown("Category", names, 0),
	}
}

func (f shopDeleteCategoryForm) Submit(submitter form.Submitter, tx *world.Tx) {
	id := f.categoryIDs[f.Category.Value()]
	if id == 0 {
		f.player.Message("Ã‚Â§cThere are no categories to delete.")
		return
	}
	if err := f.sessionManager.DeleteShopCategory(id); err != nil {
		f.player.Messagef("Ã‚Â§cUnable to delete category: %v", err)
		return
	}
	f.player.Message("Ã‚Â§aShop category deleted.")
	openShopAdminMenu(f.player, f.sessionManager)
}

type shopDeleteItemForm struct {
	sessionManager *session.Manager
	player         *player.Player
	itemIDs        []int64

	Item form.Dropdown
}

func newShopDeleteItemForm(p *player.Player, s *session.Manager) shopDeleteItemForm {
	entries := make([]shop.Entry, 0)
	for _, category := range s.GetShopCategories() {
		entries = append(entries, s.GetShopItemsByCategory(category.ID)...)
	}
	names := make([]string, 0, len(entries))
	ids := make([]int64, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Display())
		ids = append(ids, entry.ID)
	}
	if len(names) == 0 {
		names = []string{"No items"}
		ids = []int64{0}
	}
	return shopDeleteItemForm{
		sessionManager: s,
		player:         p,
		itemIDs:        ids,
		Item:           form.NewDropdown("Item", names, 0),
	}
}

func (f shopDeleteItemForm) Submit(submitter form.Submitter, tx *world.Tx) {
	id := f.itemIDs[f.Item.Value()]
	if id == 0 {
		f.player.Message("Ã‚Â§cThere are no items to delete.")
		return
	}
	if err := f.sessionManager.DeleteShopItem(id); err != nil {
		f.player.Messagef("Ã‚Â§cUnable to delete item: %v", err)
		return
	}
	f.player.Message("Ã‚Â§aShop item deleted.")
	openShopAdminMenu(f.player, f.sessionManager)
}

func newShopItemForm(p *player.Player, s *session.Manager) shopItemForm {
	categories := s.GetShopCategories()
	names := make([]string, 0, len(categories))
	ids := make([]int64, 0, len(categories))
	for _, category := range categories {
		names = append(names, category.Name)
		ids = append(ids, category.ID)
	}
	if len(names) == 0 {
		names = []string{"No categories"}
		ids = []int64{0}
	}
	return shopItemForm{
		sessionManager: s,
		player:         p,
		categoryIDs:    ids,
		Category:       form.NewDropdown("Category", names, 0),
		Identifier:     form.NewInput("Item identifier", "", "minecraft:diamond"),
		Meta:           form.NewInput("Meta", "0", "0"),
		Count:          form.NewInput("Count", "1", "1"),
		BuyPrice:       form.NewInput("Buy price", "100", "100"),
		SellPrice:      form.NewInput("Sell price", "50", "50"),
		DisplayName:    form.NewInput("Display name", "", "Diamond"),
		Lore:           form.NewInput("Lore (single line)", "", "Rare crafting gem"),
		Sort:           form.NewInput("Sort order", "10", "10"),
	}
}

func (f shopItemForm) Submit(submitter form.Submitter, tx *world.Tx) {
	if len(f.categoryIDs) == 0 || f.categoryIDs[f.Category.Value()] == 0 {
		f.player.Message("Ã‚Â§cCreate a category first.")
		return
	}
	meta, _ := strconv.Atoi(strings.TrimSpace(f.Meta.Value()))
	count, _ := strconv.Atoi(strings.TrimSpace(f.Count.Value()))
	buyPrice, _ := strconv.Atoi(strings.TrimSpace(f.BuyPrice.Value()))
	sellPrice, _ := strconv.Atoi(strings.TrimSpace(f.SellPrice.Value()))
	sortOrder, _ := strconv.Atoi(strings.TrimSpace(f.Sort.Value()))
	entry := &shop.Entry{
		CategoryID:  f.categoryIDs[f.Category.Value()],
		Identifier:  strings.TrimSpace(f.Identifier.Value()),
		Meta:        int16(meta),
		Count:       count,
		BuyPrice:    buyPrice,
		SellPrice:   sellPrice,
		DisplayName: strings.TrimSpace(f.DisplayName.Value()),
		Lore:        strings.TrimSpace(f.Lore.Value()),
		Sort:        sortOrder,
	}
	if entry.Identifier == "" || entry.Count <= 0 || entry.BuyPrice < 0 || entry.SellPrice < 0 {
		f.player.Message("Ã‚Â§cFill out valid item values before saving.")
		return
	}
	if err := f.sessionManager.SaveShopItem(entry); err != nil {
		f.player.Messagef("Ã‚Â§cUnable to save item: %v", err)
		return
	}
	f.player.Messagef("Ã‚Â§aSaved shop item '%s'.", entry.Display())
	openShopAdminMenu(f.player, f.sessionManager)
}

func openFactionMainMenu(p *player.Player, s *session.Manager) {
	buttons := []form.Button{
		form.NewButton("Faction Dashboard", "textures/ui/multiplayer_glyph_color"),
		form.NewButton("Territory", "textures/ui/world_glyph_color_2x"),
		form.NewButton("Members & Roles", "textures/ui/FriendsIcon"),
		form.NewButton("Rankings", "textures/ui/op"),
		form.NewButton("Economy & Shop", "textures/ui/icon_best3"),
		form.NewButton("Help & Commands", "textures/items/book_normal"),
	}

	if _, ok := s.GetPlayerFaction(p.UUID()); !ok {
		buttons = []form.Button{
			form.NewButton("Create Faction", "textures/ui/color_plus"),
			form.NewButton("Rankings", "textures/ui/op"),
			form.NewButton("Economy & Shop", "textures/ui/icon_best3"),
			form.NewButton("Help & Commands", "textures/items/book_normal"),
		}
	}

	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			switch pressed.Text {
			case "Faction Dashboard":
				if fac, ok := s.GetPlayerFaction(submitter.UUID()); ok {
					sendFactionInfo(fac, nil, submitter)
				} else {
					submitter.Message("§eYou are not in a faction yet. Create one to unlock faction tools.")
				}
			case "Territory":
				openFactionTerritoryMenu(submitter, s)
			case "Members & Roles":
				openFactionMemberMenu(submitter, s)
			case "Rankings":
				openFactionRankingMenu(submitter, s)
			case "Economy & Shop":
				openFactionEconomyMenu(submitter, s)
			case "Help & Commands":
				sendFactionHelpChat(submitter)
				openFactionHelpMenu(submitter)
			case "Create Faction":
				submitter.SendForm(form.New(newFactionCreateForm(submitter, s), "Create Faction"))
			}
		},
	}, "Faction Control").WithBody("Manage factions, territory, economy, and rankings from one place.").WithButtons(buttons...)
	p.SendForm(menu)
}

func openFactionTerritoryMenu(p *player.Player, s *session.Manager) {
	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			switch pressed.Text {
			case "Claim Here":
				handleFactionClaim(submitter, nil, s)
			case "Unclaim Here":
				handleFactionUnclaim(submitter, nil, s)
			case "Toggle Borders":
				handleFactionBorderToggle(submitter, nil, s)
			case "Claim Map":
				openFactionClaimMapMenu(submitter, s, nil)
			case "Back":
				openFactionMainMenu(submitter, s)
			}
		},
	}, "Territory").WithBody("Claim land, review nearby chunks, and manage border visibility.").WithButtons(
		form.NewButton("Claim Here", "textures/ui/realms_green_check"),
		form.NewButton("Unclaim Here", "textures/ui/icon_trash"),
		form.NewButton("Toggle Borders", "textures/ui/world_glyph_color_2x"),
		form.NewButton("Claim Map", "textures/items/map_filled"),
		form.NewButton("Back", "textures/ui/cancel"),
	)
	p.SendForm(menu)
}

func openFactionMemberMenu(p *player.Player, s *session.Manager) {
	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			switch pressed.Text {
			case "View Members":
				FactionWho{sessionManager: s}.Run(submitter, nil, tx)
			case "Role Commands":
				submitter.Message("§ePromote: /f promote <player>")
				submitter.Message("§eDemote: /f demote <player>")
				submitter.Message("§eKick: /f kick <player>")
			case "Description Help":
				submitter.Message("§eUpdate faction description with: /f desc <text>")
			case "Back":
				openFactionMainMenu(submitter, s)
			}
		},
	}, "Members & Roles").WithBody("View your roster and use management commands for ranks and descriptions.").WithButtons(
		form.NewButton("View Members", "textures/ui/FriendsIcon"),
		form.NewButton("Role Commands", "textures/items/book_normal"),
		form.NewButton("Description Help", "textures/ui/editIcon"),
		form.NewButton("Back", "textures/ui/cancel"),
	)
	p.SendForm(menu)
}

func openFactionRankingMenu(p *player.Player, s *session.Manager) {
	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			switch pressed.Text {
			case "Top Factions":
				openTopFactionsMenu(submitter, s)
			case "Top Players":
				openTopPlayersMenu(submitter, s)
			case "Back":
				openFactionMainMenu(submitter, s)
			}
		},
	}, "Rankings").WithBody("Browse power rankings for factions and players.").WithButtons(
		form.NewButton("Top Factions", "textures/ui/op"),
		form.NewButton("Top Players", "textures/ui/friend_glyph"),
		form.NewButton("Back", "textures/ui/cancel"),
	)
	p.SendForm(menu)
}

func openFactionEconomyMenu(p *player.Player, s *session.Manager) {
	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			switch pressed.Text {
			case "Bounties":
				openBountyMenu(submitter, s)
			case "Faction Shop":
				openShopCategoriesMenu(submitter, s)
			case "Balance Help":
				data := s.GetOrCreatePlayer(submitter.UUID(), submitter.Name())
				submitter.Messagef("§aCurrent balance: §f$%d", data.Balance)
				submitter.Message("§eUse /pay <player> <amount> to send money.")
			case "Back":
				openFactionMainMenu(submitter, s)
			}
		},
	}, "Economy & Shop").WithBody("Open the shop, review bounties, and manage your money flow.").WithButtons(
		form.NewButton("Bounties", "textures/ui/icon_import"),
		form.NewButton("Faction Shop", "textures/ui/icon_best3"),
		form.NewButton("Balance Help", "textures/items/gold_ingot"),
		form.NewButton("Back", "textures/ui/cancel"),
	)
	p.SendForm(menu)
}

func openBountyMenu(p *player.Player, s *session.Manager) {
	lines := []string{"Top bounties:"}
	for i, data := range s.TopBounties(5) {
		if data.Bounty <= 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("Ã‚Â§6#%d Ã‚Â§f%s Ã‚Â§7- Ã‚Â§c$%d", i+1, data.Name, data.Bounty))
	}
	body := strings.Join(lines, "\n")
	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			switch pressed.Text {
			case "Place Bounty":
				submitter.SendForm(form.New(newBountyPlaceForm(submitter, s), "Place Bounty"))
			case "Refresh":
				openBountyMenu(submitter, s)
			}
		},
	}, "Bounty Board").WithBody(body).WithButtons(
		form.NewButton("Place Bounty", "textures/ui/color_plus"),
		form.NewButton("Refresh", "textures/ui/refresh_light"),
	)
	p.SendForm(menu)
}

func openShopCategoriesMenu(p *player.Player, s *session.Manager) {
	categories := s.GetShopCategories()
	buttons := make([]form.Button, 0, len(categories))
	for _, category := range categories {
		buttons = append(buttons, form.NewButton(category.Name, category.Icon))
	}
	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			for _, category := range categories {
				if category.Name == pressed.Text {
					openShopItemsMenu(submitter, s, category.ID)
					return
				}
			}
		},
	}, "Faction Shop").WithBody("Buy and sell items using your balance.").WithButtons(buttons...)
	p.SendForm(menu)
}

func openShopItemsMenu(p *player.Player, s *session.Manager, categoryID int64) {
	category, ok := s.GetShopCategory(categoryID)
	if !ok {
		p.Message("Ã‚Â§cThat shop category no longer exists.")
		return
	}
	items := s.GetShopItemsByCategory(categoryID)
	buttons := make([]form.Button, 0, len(items))
	for _, entry := range items {
		label := fmt.Sprintf("%s\nÃ‚Â§aBuy $%d Ã‚Â§7/ Ã‚Â§6Sell $%d", entry.Display(), entry.BuyPrice, entry.SellPrice)
		buttons = append(buttons, form.NewButton(label, category.Icon))
	}

	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			for _, entry := range items {
				if strings.HasPrefix(pressed.Text, entry.Display()) {
					openShopEntryMenu(submitter, s, entry)
					return
				}
			}
		},
	}, category.Name).WithBody("Select an item to buy or sell.").WithButtons(buttons...)
	p.SendForm(menu)
}

func openShopEntryMenu(p *player.Player, s *session.Manager, entry shop.Entry) {
	body := fmt.Sprintf("Ã‚Â§fItem: Ã‚Â§b%s\nÃ‚Â§fBuy: Ã‚Â§a$%d\nÃ‚Â§fSell: Ã‚Â§6$%d\nÃ‚Â§fCount: Ã‚Â§e%d", entry.Display(), entry.BuyPrice, entry.SellPrice, entry.Count)
	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			switch pressed.Text {
			case "Buy":
				if err := s.BuyShopItem(submitter.UUID(), entry.ID); err != nil {
					submitter.Messagef("Ã‚Â§cBuy failed: %v", err)
					return
				}
				submitter.Messagef("Ã‚Â§aBought %s for $%d.", entry.Display(), entry.BuyPrice)
			case "Sell":
				if err := s.SellShopItem(submitter.UUID(), entry.ID); err != nil {
					submitter.Messagef("Ã‚Â§cSell failed: %v", err)
					return
				}
				submitter.Messagef("Ã‚Â§aSold %s for $%d.", entry.Display(), entry.SellPrice)
			}
		},
	}, entry.Display()).WithBody(body).WithButtons(
		form.NewButton("Buy", "textures/ui/confirm"),
		form.NewButton("Sell", "textures/ui/trade_icon"),
	)
	p.SendForm(menu)
}

func openShopAdminMenu(p *player.Player, s *session.Manager) {
	menu := form.NewMenu(menuSubmit{
		onSubmit: func(submitter *player.Player, pressed form.Button, tx *world.Tx) {
			switch pressed.Text {
			case "Add Category":
				submitter.SendForm(form.New(newShopCategoryForm(submitter, s), "Add Shop Category"))
			case "Add Item":
				submitter.SendForm(form.New(newShopItemForm(submitter, s), "Add Shop Item"))
			case "Delete Category":
				submitter.SendForm(form.New(newShopDeleteCategoryForm(submitter, s), "Delete Shop Category"))
			case "Delete Item":
				submitter.SendForm(form.New(newShopDeleteItemForm(submitter, s), "Delete Shop Item"))
			case "List Categories":
				openShopCategoriesMenu(submitter, s)
			}
		},
	}, "Shop Admin").WithBody("Create categories and add configurable shop entries.").WithButtons(
		form.NewButton("Add Category", "textures/ui/color_plus"),
		form.NewButton("Add Item", "textures/ui/icon_import"),
		form.NewButton("Delete Category", "textures/ui/icon_trash"),
		form.NewButton("Delete Item", "textures/ui/icon_delete"),
		form.NewButton("List Categories", "textures/ui/book_edit_default"),
	)
	p.SendForm(menu)
}

func parseUUID(v string) (uuid.UUID, error) {
	id, err := strconv.Unquote(`"` + v + `"`)
	if err == nil {
		v = id
	}
	return uuid.Parse(v)
}

