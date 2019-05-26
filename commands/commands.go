package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/utils"
)

const (
	MESSAGE_LIMIT = 2000
	SEND_LIMIT    = 10
)

// Command The command interface that all commands implement
type Command interface {
	Names() []string
	Desc() string
	Roles() []string
	Chans() []string

	MsgHandle(*discordgo.Session, *discordgo.Message, []string) (*CommandSend, error)
}

// Send Stores the stuff we need to send
type CommandSend struct {
	data      []*discordgo.MessageSend
	channelid string
}

// NewSend Returns a send struct.
func NewSend(cid string) *CommandSend {
	return &CommandSend{
		make([]*discordgo.MessageSend, SEND_LIMIT),
		cid,
	}
}

// NewSimpleSend Returns a send struct with the message content filled in.
func NewSimpleSend(cid string, msg string) *CommandSend {
	send := &discordgo.MessageSend{
		Content: msg,
		Embed:   nil,
		Tts:     false,
		Files:   nil,
		File:    nil,
	}
	return &CommandSend{
		data:      []*discordgo.MessageSend{send},
		channelid: cid,
	}
}

// AddSimpleMessage Adds another simple message to be sent.
func (c *CommandSend) AddSimpleMessage(msg string) {
	send := &discordgo.MessageSend{
		Content: msg,
		Embed:   nil,
		Tts:     false,
		Files:   nil,
		File:    nil,
	}
	c.data = append(c.data, send)
}

// AddEmbedMessage Adds an embed message to be sent.
func (c *CommandSend) AddEmbedMessage(emb *discordgo.MessageEmbed) {
	send := &discordgo.MessageSend{
		Content: "",
		Embed:   emb,
		Tts:     false,
		Files:   nil,
		File:    nil,
	}
	c.data = append(c.data, send)
}

// AddMessageSend Adds a discordgo MessageSend
func (c *CommandSend) AddMessageSend(send *discordgo.MessageSend) {
	c.data = append(c.data, send)
}

// Send Sends the messages a command returns while also checking message length
func (c *CommandSend) Send(s *discordgo.Session) error {
	// Get the stuff out of BeegYoshi and send it into the server
	for _, data := range c.data {
		if utils.Strlen(data) > MESSAGE_LIMIT {
			return fmt.Errorf("Send: following message exceeds limit\n%#v", data)
		}
		s.ChannelMessageSendComplex(c.channelid, data)
	}
	return nil
}
