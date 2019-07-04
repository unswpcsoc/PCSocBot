package handlers

import (
	"github.com/bwmarrin/discordgo"
	. "github.com/unswpcsoc/PCSocBot/commands"
)

type Ping struct{}

func NewPing() *Ping { return &Ping{} }

func (p *Ping) Aliases() []string { return []string{"ping", "ping pong"} }

func (p *Ping) Desc() string { return "Ping!" }

func (p *Ping) Roles() []string { return nil }

func (p *Ping) Chans() []string { return nil }

func (p *Ping) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	return NewSimpleSend(msg.ChannelID, "Pong!"), nil
}
