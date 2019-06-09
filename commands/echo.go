package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Echo struct {
	names []string
	desc  string
}

func NewEcho() *Echo {
	return &Echo{
		names: []string{"echo"},
		desc:  "Echo!",
	}
}

func (e *Echo) Names() []string {
	return e.names
}

func (e *Echo) Desc() string {
	return e.desc
}

func (e *Echo) Roles() []string {
	return nil
}

func (e *Echo) Chans() []string {
	return nil
}

func (e *Echo) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	var out string
	if len(args) == 0 {
		out = "Echo!"
	} else {
		out = strings.Join(args, " ")
	}

	return NewSimpleSend(msg.ChannelID, out), nil
}
