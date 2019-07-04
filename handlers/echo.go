package handlers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	. "github.com/unswpcsoc/PCSocBot/commands"
)

type Echo struct {
	Input []string `arg:"input"`
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

func (e *Echo) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	var out string
	if len(e.Input) == 0 {
		out = "Echo!"
	} else {
		out = strings.Join(e.Input, " ")
	}

	return NewSimpleSend(msg.ChannelID, out), nil
}
