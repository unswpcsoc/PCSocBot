// Package handlers contains concrete implementations of the Command interface
//
package handlers

import (
	"github.com/bwmarrin/discordgo"

	. "github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/router"
)

var crt *router.Router

func init() {
	crt = router.NewRouter()

	crt.AddCommand(NewDecimalSpiral())

	crt.AddCommand(NewEcho())

	crt.AddCommand(NewHelp())

	crt.AddCommand(NewLog())

	crt.AddCommand(NewPing())

	crt.AddCommand(NewQuote())
	crt.AddCommand(NewQuoteAdd())
	crt.AddCommand(NewQuoteApprove())
	crt.AddCommand(NewQuoteList())
	crt.AddCommand(NewQuotePending())
	crt.AddCommand(NewQuoteRemove())
	crt.AddCommand(NewQuoteReject())

	crt.AddCommand(NewRole("Bookworm"))
	crt.AddCommand(NewRole("Meta"))
	crt.AddCommand(NewRole("Weeb"))

	crt.AddCommand(NewTags())
	crt.AddCommand(NewTagsAdd())
	crt.AddCommand(NewTagsClean())
	crt.AddCommand(NewTagsList())
	crt.AddCommand(NewTagsGet())
	crt.AddCommand(NewTagsPing())
	crt.AddCommand(NewTagsPingMe())
	crt.AddCommand(NewTagsPlatforms())
	crt.AddCommand(NewTagsRemove())
	crt.AddCommand(NewTagsUser())

	crt.AddCommand(NewArchive())
}

// RouterRoute is a wrapper around the handler package's internal router's Route method
func RouterRoute(argv []string) (Command, int) {
	return crt.Route(argv)
}

// RouterToSlice is a wrapper around the blah blah blah's ToSlice method
func RouterToSlice() []Command {
	return crt.ToSlice()
}

// NilCommand is a thing that you can struct embed to avoid boilerplate
type NilCommand struct{}

func (n *NilCommand) Aliases() []string { return []string{""} }

func (n *NilCommand) Desc() string { return "" }

func (n *NilCommand) Subcommands() []Command { return nil }

func (n *NilCommand) Roles() []string { return nil }

func (n *NilCommand) Chans() []string { return nil }

// InitLogs inits all logging commands
func InitLogs(ses *discordgo.Session) {
	initFill(ses)
	initDel(ses)
	initArchive(ses)
}
