package handlers

import (
	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/router"
)

var commandRouter router.Router

func init() {
	commandRouter = router.NewRouter()

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
}

// Route is a wrapper around the handler package's internal router's Route method
func Route(argv []string) (commands.Command, int) {
	return commandRouter.Route(argv)
}
