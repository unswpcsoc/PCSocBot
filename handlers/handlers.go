// Package containing concrete implementations of the Command interface
//
// Boilerplate:
//  import (
//  	"github.com/bwmarrin/discordgo"
//  	. "github.com/unswpcsoc/PCSocBot/commands"
//  )
//
//  type YourCommand struct{}
//
//  func NewYourCommand() *YourCommand { return &YourCommand{} }
//
//  func ( *YourCommand) Aliases() []string { return []string{"",}
//
//  func ( *YourCommand) Desc() string { return "YourCommand!" }
//
//  func ( *YourCommand) Roles() []string { return nil }
//
//  func ( *YourCommand) Chans() []string { return nil }
//
//  func ( *YourCommand) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) { return nil, nil }
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
	commandRouter.Addcommand(NewTagsList())
	commandRouter.Addcommand(NewTagsGet())
	commandRouter.Addcommand(NewTagsPing())
	commandRouter.Addcommand(NewTagsPingMe())
	commandRouter.Addcommand(NewTagsPlatforms())
	commandRouter.Addcommand(NewTagsRemove())
	commandRouter.Addcommand(NewTagsUser())
}

// Route is a wrapper around the handler package's internal router's Route method
func Route(argv []string) (Command, int) {
	return commandRouter.Route(argv)
}

// ToSlice is a wrapper around the blah blah blah's ToSlice method
func ToSlice() []Command {
	return commandRouter.ToSlice()
}
