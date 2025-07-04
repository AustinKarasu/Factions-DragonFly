package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/jorgebyte/faction/internal/session"
)

func RegisterAll(s *session.Manager) {
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

	cmd.Register(cmd.New("f", "Command Factions", []string{"fac"},
		createCmd,
		infoSelfCmd,
		infoOtherCmd,
		deleteCmd,
		inviteCmd,
		joinCmd,
		leaveCmd,
		topCmd,
		allyCmd,
		claimCmd,
		unclaimCmd,
		borderCmd,
	))
}
