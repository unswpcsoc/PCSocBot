package commands

import (
	"github.com/bwmarrin/discordgo"
)

// Command The command
type Command interface {
	Names() []string
	Desc() string
	Roles() []string
	Channels() []string

	MsgHandle() (Send, error)
}

// Send Stores the stuff we need to send
type Send struct {
	BeegYoshi map[int]discordgo.MessageSend
}

func (s *Send) AddChungus() {
	// Get the stuff out of BeegYoshi and send it into the server

	...

	// Other stuff to do (eg: add reacts probably done in Main)
	// ()
}
