// Package containing concrete implementations of the Command interface
//
package handlers

import (
	. "github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/router"
)

var crt *router.Router

func init() {
	crt = router.NewRouter()

	crt.Addcommand(NewDecimalSpiral())

	crt.Addcommand(NewEcho())

	crt.Addcommand(NewHelp())

	crt.Addcommand(NewLog())

	crt.Addcommand(NewPing())

	crt.Addcommand(NewQuote())
	crt.Addcommand(NewQuoteAdd())
	crt.Addcommand(NewQuoteApprove())
	crt.Addcommand(NewQuoteList())
	crt.Addcommand(NewQuotePending())
	crt.Addcommand(NewQuoteRemove())
	crt.Addcommand(NewQuoteReject())

	crt.Addcommand(NewRole("Bookworm"))
	crt.Addcommand(NewRole("Meta"))
	crt.Addcommand(NewRole("Weeb"))

	crt.Addcommand(NewTags())
	crt.Addcommand(NewTagsAdd())
	crt.Addcommand(NewTagsClean())
	crt.Addcommand(NewTagsList())
	crt.Addcommand(NewTagsGet())
	crt.Addcommand(NewTagsPing())
	crt.Addcommand(NewTagsPingMe())
	crt.Addcommand(NewTagsPlatforms())
	crt.Addcommand(NewTagsRemove())
	crt.Addcommand(NewTagsUser())
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
