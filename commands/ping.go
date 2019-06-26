package commands

import "github.com/bwmarrin/discordgo"

type Ping struct{}

func NewPing() *Ping { return &Ping{} }

func (p *Ping) Aliases() []string { return []string{"ping", "ping pong"} }

func (p *Ping) Desc() string { return "Ping!" }

func (p *Ping) Roles() []string { return nil }

func (p *Ping) Chans() []string { return nil }

func (p *Ping) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	return NewSimpleSend(msg.ChannelID, "Pong!"), nil
}
