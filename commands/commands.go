package commands

import (
	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/utils"
)

// Command The command
type Command interface {
	Names() []string
	Desc() string
	Roles() []string
	Channels() []string

	MsgHandle(*discordgo.Session, *discordgo.MessageCreate) (Send, error)
}

// Send Stores the stuff we need to send
type CommandSend struct {
	data      []*discordgo.MessageSend
	channelid string
}

// NewSend Returns a send struct.
func NewSend(cid string) *CommandSend {
	return &CommandSend{
		make([]*discordgo.MessageSend),
		cid,
	}
}

// NewSimpleSend Returns a send struct with the message content filled in.
func NewSimpleSend(cid string, msg string) *CommandSend {
	send := discordgo.MessageSend{
		Content: msg,
		nil,
		nil,
		nil,
		nil,
	}
	return &CommandSend{
		data:      []*discordgo.MessageSend{send},
		channelid: cid,
	}
}

// AddSimpleMessage Adds another simple message to be sent.
func (c *CommandSend) AddSimpleMessage(msg string) {
	send := discordgo.MessageSend{
		Content: msg,
		nil,
		nil,
		nil,
		nil,
	}
	append(c.data, send)
}

// AddEmbedMessage Adds an embed message to be sent.
func (c *CommandSend) AddEmbedMessage(emb *discordgo.MessageEmbed) {
	send := discordgo.MessageSend{
		nil,
		emb,
		nil,
		nil,
		nil,
	}
	append(c.data, send)
}

// AddMessageSend Adds a pure message embed to be sent.
func (c *CommandSend) AddMessageSend(msd *discordgo.MessageSend) {
	append(c.data, send)
}

// Send Sends the messages a command returns while also checking message length
func (c *CommandSend) Send(s *discordgo.Session) {
	// Get the stuff out of BeegYoshi and send it into the server
	// TODO: Check message length
	for _, data := range c.data {
		s.ChannelMessageSendComplex(c.channelid, data)
	}
}
