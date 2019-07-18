// Package containing concrete implementations of the Command interface
//
package handlers

import (
	. "github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/router"
)

var commandRouter *router.Router

func init() {
	commandRouter = router.NewRouter()
	commandRouter.Addcommand(NewHelp())

	commandRouter.Addcommand(NewPing())

	commandRouter.Addcommand(NewEcho())

	commandRouter.Addcommand(NewQuote())
	commandRouter.Addcommand(NewQuoteAdd())
	commandRouter.Addcommand(NewQuoteApprove())
	commandRouter.Addcommand(NewQuoteList())
	commandRouter.Addcommand(NewQuotePending())
	commandRouter.Addcommand(NewQuoteRemove())
	commandRouter.Addcommand(NewQuoteReject())

	commandRouter.Addcommand(NewDecimalSpiral())

	commandRouter.Addcommand(NewRole("Bookworm"))
	commandRouter.Addcommand(NewRole("Meta"))
	commandRouter.Addcommand(NewRole("Weeb"))

	commandRouter.Addcommand(NewTags())
	commandRouter.Addcommand(NewTagsAdd())
	commandRouter.Addcommand(NewTagsClean())
	commandRouter.Addcommand(NewTagsList())
	commandRouter.Addcommand(NewTagsGet())
	commandRouter.Addcommand(NewTagsPing())
	commandRouter.Addcommand(NewTagsPingMe())
	commandRouter.Addcommand(NewTagsPlatforms())
	commandRouter.Addcommand(NewTagsRemove())
	commandRouter.Addcommand(NewTagsUser())
}

// RouterRoute is a wrapper around the handler package's internal router's Route method
func RouterRoute(argv []string) (Command, int) {
	return commandRouter.Route(argv)
}

// RouterToSlice is a wrapper around the blah blah blah's ToSlice method
func RouterToSlice() []Command {
	return commandRouter.ToSlice()
}

// NilCommand is a thing that you can struct embed to avoid boilerplate
type NilCommand struct{}

func (n *NilCommand) Aliases() []string { return []string{""} }

func (n *NilCommand) Desc() string { return "" }

func (n *NilCommand) Subcommands() []Command { return nil }

func (n *NilCommand) Roles() []string { return nil }

func (n *NilCommand) Chans() []string { return nil }
