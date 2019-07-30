package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/unswpcsoc/PCSocBot/commands"
)

type ping struct {
	nilCommand
}

func newPing() *ping { return &ping{} }

func (p *ping) Aliases() []string { return []string{"ping", "ping pong"} }

func (p *ping) Desc() string { return "ping!" }

func (p *ping) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	return commands.NewSimpleSend(msg.ChannelID, "Pong!"), nil
}
