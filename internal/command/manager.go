package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/jorgebyte/faction/internal/session"
)

func RegisterAll(s *session.Manager) {
	menuCmd := FactionMenu{sessionManager: s}
	createCmd := FactionCreate{sessionManager: s}
	infoSelfCmd := FactionInfoSelf{sessionManager: s}
	infoOtherCmd := FactionInfoOther{sessionManager: s}
	deleteCmd := FactionDelete{sessionManager: s}
	inviteCmd := FactionInvite{sessionManager: s}
	joinCmd := FactionJoin{sessionManager: s}
	leaveCmd := FactionLeave{sessionManager: s}
	topCmd := FactionTop{sessionManager: s}
	allyCmd := FactionAlly{sessionManager: s}
	claimCmd := FactionClaim{sessionManager: s}
	unclaimCmd := FactionUnclaim{sessionManager: s}
	borderCmd := FactionBorder{sessionManager: s}
	helpCmd := FactionHelp{sessionManager: s}
	overviewCmd := FactionOverview{sessionManager: s}
	whoCmd := FactionWho{sessionManager: s}
	mapCmd := FactionMap{sessionManager: s}

	cmd.Register(cmd.New("f", "Command Factions", []string{"fac"},
		menuCmd,
		helpCmd,
		createCmd,
		infoSelfCmd,
		infoOtherCmd,
		overviewCmd,
		whoCmd,
		deleteCmd,
		inviteCmd,
		joinCmd,
		leaveCmd,
		topCmd,
		allyCmd,
		claimCmd,
		unclaimCmd,
		borderCmd,
		mapCmd,
	))

	cmd.Register(cmd.New("shop", "Open the faction shop", nil,
		ShopOpen{sessionManager: s},
		ShopAdmin{sessionManager: s},
	))

	cmd.Register(cmd.New("bounty", "Manage player bounties", nil,
		BountyOpen{sessionManager: s},
		BountyTop{sessionManager: s},
		BountySet{sessionManager: s},
		BountyClear{sessionManager: s},
	))

	cmd.Register(cmd.New("balance", "View your balance", []string{"bal", "money"},
		Balance{sessionManager: s},
	))

	cmd.Register(cmd.New("pay", "Pay another player", nil,
		Pay{sessionManager: s},
	))
}
