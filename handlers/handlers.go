// Package handlers contains concrete implementations of the Command interface
//
package handlers

import (
	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/internal/router"
)

var crt *router.Router

func init() {
	crt = router.NewRouter()

	crt.AddCommand(newDecimalSpiral())

	crt.AddCommand(newEcho())

	crt.AddCommand(newHelp())

	crt.AddCommand(newLog())
	crt.AddCommand(newLogDelete())
	crt.AddCommand(newLogFilter())

	crt.AddCommand(newPing())

	crt.AddCommand(newQuote())
	crt.AddCommand(newQuoteAdd())
	crt.AddCommand(newQuoteApprove())
	crt.AddCommand(newQuoteList())
	crt.AddCommand(newQuotePending())
	crt.AddCommand(newQuoteRemove())
	crt.AddCommand(newQuoteReject())
	crt.AddCommand(newQuoteSearch())

	crt.AddCommand(newRole("Bookworm"))
	crt.AddCommand(newRole("Meta"))
	crt.AddCommand(newRole("Weeb"))

	crt.AddCommand(newTags())
	crt.AddCommand(newTagsAdd())
	crt.AddCommand(newTagsClean())
	crt.AddCommand(newTagsGet())
	crt.AddCommand(newTagsList())
	crt.AddCommand(newTagsModRemove())
	crt.AddCommand(newTagsPing())
	crt.AddCommand(newTagsPingMe())
	crt.AddCommand(newTagsPlatforms())
	crt.AddCommand(newTagsRemove())
	crt.AddCommand(newTagsUser())

	crt.AddCommand(newArchive())

	crt.AddCommand(newStaticIce())

	crt.AddCommand(newHandbook())

	crt.AddCommand(newScream())

}

// RouterRoute is a wrapper around the handler package's internal router's Route method
func RouterRoute(argv []string) (commands.Command, int) { return crt.Route(argv) }

// RouterToSlice is a wrapper around the blah blah blah's ToSlice method
func RouterToSlice() []commands.Command { return crt.ToSlice() }

// RouterToStringSlice is a wrapper around the blah blah blah's ToStringSlice method
func RouterToStringSlice() []string { return crt.ToStringSlice() }

// nilCommand is a thing that you can struct embed to avoid boilerplate
type nilCommand struct{}

func (n *nilCommand) Subcommands() []commands.Command { return nil }

func (n *nilCommand) Roles() []string { return nil }

func (n *nilCommand) Chans() []string { return nil }

// InitLogs inits all logging commands.
// Needs to be maually updated when adding new loggers
func InitLogs(ses *discordgo.Session) {
	initFil(ses)
	initDel(ses)
	initArchive(ses)
}
