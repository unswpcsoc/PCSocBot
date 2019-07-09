package handlers

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	. "github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/utils"
)

const HELPALIAS = PREFIX + "h"

// Help is a special command that needs a concrete router to work
type Help struct {
	Query []string `arg:"query"`
}

func NewHelp() *Help { return &Help{} }

func (h *Help) Aliases() []string { return []string{"helpme", "h", "commands", "fuck", "fuck you"} }

func (h *Help) Desc() string { return "Help!" }

func (h *Help) Subcommands() []Command { return nil }

func (h *Help) Roles() []string { return nil }

func (h *Help) Chans() []string { return nil }

func (h *Help) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	snd := NewSend(msg.ChannelID)
	if len(h.Query) == 0 {
		// consider rate-limiting/re-routing/disabling this if your message is enormous
		count := 0
		out := utils.Bold("All Commands:")
		for _, com := range ToSlice() {
			tmp := "\n" + GetUsage(com)
			count += len(tmp)
			if count < 2000 {
				out += tmp
			} else {
				snd.AddSimpleMessage(out)
				out = tmp
				count = len(tmp)
			}
		}
		snd.AddSimpleMessage(out)
	} else {
		com, _ := Route(h.Query)
		if com == nil {
			return nil, errors.New("Error: Unknown command; use " + HELPALIAS)
		}
		snd.AddSimpleMessage(GetUsage(com))
	}
	return snd, nil
}
