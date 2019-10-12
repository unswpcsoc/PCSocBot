package handlers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sahilm/fuzzy"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/internal/utils"
)

const (
	// HelpAlias is the default alias for help command
	HelpAlias = "hg"
)

// help is a special command that needs a concrete router to work
type help struct {
	nilCommand
	Query []string `arg:"query"`
}

func newHelp() *help { return &help{} }

func (h *help) Aliases() []string { return []string{HelpAlias, "commands", "fuck", "fuck you"} }

func (h *help) Desc() string { return "help!" }

func (h *help) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	snd := commands.NewSend(msg.ChannelID)
	var out string
	if len(h.Query) == 0 {
		// consider rate-limiting/re-routing/disabling this if your message becomes enormous
		ignore := map[string]bool{}
		routerSlice := RouterToSlice()
		for _, com := range routerSlice {
			// register subcommands to be ignored using first alias
			if com.Subcommands() != nil {
				for _, sub := range com.Subcommands() {
					// check if programmer has accidentally included root command as a subcommand
					if sub.Aliases()[0] == com.Aliases()[0] {
						panic("you idiot")
					}

					ignore[sub.Aliases()[0]] = true
				}
			}
		}

		count := 0
		out = utils.Bold("All Commands:")
		for _, com := range routerSlice {
			// ignore subcommands
			if seen, _ := ignore[com.Aliases()[0]]; seen {
				continue
			}

			tmp := "\n" + commands.GetUsage(com)
			count += len(tmp)
			if count < commands.MessageLimit {
				out += tmp
			} else {
				snd.Message(out)
				out = tmp
				count = len(tmp)
			}
		}
		snd.Message(out)
	} else {
		com, _ := RouterRoute(h.Query)
		if com != nil {
			out = "Command " + utils.Bold(com.Aliases()[0])
			out += "\n" + commands.GetUsage(com)
			snd.Message(out)
		} else {
			// user provided bad command string, use fuzzy finding to find suggestions
			mat := fuzzy.Find(strings.Join(h.Query, " "), RouterToStringSlice())

			out = "Unknown command provided"
			if len(mat) > 0 {
				out += ", did you mean:\n"

				// fuzzy find top 3 suggestions
				for i, m := range mat {
					if i == 3 {
						break
					}
					out += utils.Code(commands.Prefix+m.Str) + "\n"
				}
			}
			snd.Message(out)
		}
	}
	return snd, nil
}
