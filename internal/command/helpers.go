package command

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
)

func stringsJoin(lines []string) string {
	return strings.Join(lines, "\n")
}

func reply(p *player.Player, o *cmd.Output, msg string, args ...any) {
	if o != nil {
		if len(args) > 0 {
			o.Printf(msg, args...)
		} else {
			o.Printf(msg)
		}
	}
	if p != nil {
		if len(args) > 0 {
			p.Messagef(msg, args...)
		} else {
			p.Message(msg)
		}
	}
}

func replyErr(p *player.Player, o *cmd.Output, msg string, args ...any) {
	if o != nil {
		if len(args) > 0 {
			o.Errorf(msg, args...)
		} else {
			o.Errorf(msg)
		}
	}
	if p != nil {
		if len(args) > 0 {
			p.Messagef(msg, args...)
		} else {
			p.Message(msg)
		}
	}
}
