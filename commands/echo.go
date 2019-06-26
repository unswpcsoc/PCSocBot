package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Echo struct {
	input []string `arg:"input"`
}

func NewEcho() *Echo { return &Echo{} }

func (e *Echo) Aliases() []string { return []string{"echo"} }

func (e *Echo) Desc() string { return "Echo!" }

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
