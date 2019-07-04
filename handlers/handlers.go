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
	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/router"
)

var commandRouter *router.Router

func init() {
	commandRouter = router.NewRouter()
	commandRouter.Addcommand(NewHelp(commandRouter))

	commandRouter.Addcommand(NewPing())

	commandRouter.Addcommand(NewEcho())

	commandRouter.Addcommand(NewQuote())
	commandRouter.Addcommand(NewQuoteList())
	commandRouter.Addcommand(NewQuotePending())
	commandRouter.Addcommand(NewQuoteAdd())
	commandRouter.Addcommand(NewQuoteApprove())
	commandRouter.Addcommand(NewQuoteRemove())
	commandRouter.Addcommand(NewQuoteReject())

	commandRouter.Addcommand(NewDecimalSpiral())

	commandRouter.Addcommand(NewRole("Weeb"))
	commandRouter.Addcommand(NewRole("Meta"))
	commandRouter.Addcommand(NewRole("Bookworm"))

	commandRouter.Addcommand(NewTags())
	commandRouter.Addcommand(NewTagsAdd())
	commandRouter.Addcommand(NewTagsRemove())
	commandRouter.Addcommand(NewTagsView())
	commandRouter.Addcommand(NewTagsList())
	commandRouter.Addcommand(NewTagsPlatforms())
	commandRouter.Addcommand(NewTagsGet())
	commandRouter.Addcommand(NewTagsPing())
	commandRouter.Addcommand(NewTagsPingMe())
}

// Route is a wrapper around the handler package's internal router's Route method
func Route(argv []string) (commands.Command, int) {
	return commandRouter.Route(argv)
}
