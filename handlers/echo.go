package handlers

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/commands"
)

type echo struct {
	nilCommand
	Input []string `arg:"input"`
}

func newEcho() *echo { return &echo{} }

func (e *echo) Aliases() []string { return []string{"echo"} }

func (e *echo) Desc() string { return "Echo!" }

func (e *echo) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var out string
	if len(e.Input) == 0 {
		out = "Echo!"
	} else {
		out = strings.Join(e.Input, " ")
	}

	return commands.NewSimpleSend(msg.ChannelID, out), nil
}
