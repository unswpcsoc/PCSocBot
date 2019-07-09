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
		// consider rate-limiting/re-routing/disabling this if your message becomes enormous
		ignore := map[string]bool{}
		slice := RouterToSlice()
		for _, com := range slice {
			// register subcommands to be ignored using first alias
			if com.Subcommands() != nil {
				for _, sub := range com.Subcommands() {
					ignore[sub.Aliases()[0]] = true
				}
			}
		}

		count := 0
		out := utils.Bold("All Commands:")
		for _, com := range slice {
			// ignore subcommands
			if seen, _ := ignore[com.Aliases()[0]]; seen {
				continue
			}

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
		com, _ := RouterRoute(h.Query)
		if com == nil {
			return nil, errors.New("Error: Unknown command; use " + HELPALIAS)
		}

		out := "Command " + utils.Bold(com.Aliases()[0])
		if com.Subcommands() != nil {
			out += "\n" + GetUsage(com)
			out += "\n\nSubcommands:"
			for _, sub := range com.Subcommands() {
				out += "\n" + GetUsage(sub)
			}
		} else {
			out += "\n" + GetUsage(com)
		}
		snd.AddSimpleMessage(out)
	}
	return snd, nil
}
