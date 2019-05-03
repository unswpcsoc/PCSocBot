package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	com "github.com/unswpcsoc/PCSocBot/commands"
)

var (
	dg     *discordgo.Session
	router Router
	prefix = "!"
)

// leaf A leaf of the command router tree
type leaf struct {
	command *com.Command
	leaves  map[string]*leaf
}

// NewLeaf Returns a leaf with the command
func NewLeaf(cm *com.Command) *leaf {
	return &leaf{
		command: cm,
		leaves:  make(map[string]*leaf),
	}
}

// Router Routes a command string to a command
type Router struct {
	routes *leaf
}

// AddCommand Adds command-string mapping
func (r *Router) AddCommand(cm *com.Command, cmdstr []string) {
	// AddCommand(PingTagPeople, {"ask"})
	// AddCommand(PingTagPeople, {"tags", "ping"})
	if cm == nil || len(cmdstr) == 0 {
		return
	}
	if r.routes != nil {
		// Search all known leaves
		curr := r.routes
		for {
			curr, found = curr.leaves[cmdstr[0]]
			if !found {
				break
			}
			cmdstr = cmdstr[1:]
		}
	}
	// Add new leaves
	for len(cmdstr) > 0 {
		curr.leaves[cmdstr[0]] = NewLeaf(nil)
		curr = curr.leaves[cmdstr[0]]
		cmdstr = cmdstr[1:]
	}
	// Assign command to the final leaf
	curr.command = cm
}

/*
 * Route Routes to handler from string.
 * Returns the command and an empty slice if found.
 * Returns nil and the slice of matched words to use in command assistance.
 */
func (r *Router) Route(argv []string) (*com.Command, []string) {
	if r.routes == nil {
		return nil, false
	}
	if len(argv) == 0 {
		return nil, false
	}
	curr := r.routes
	for len(argv) > 0 {
		curr, found := curr[argv[0]]
		if !found {
			break
		}
		argv = argv[1:]
	}
	/*
		cm, _ := doRoute(r.routes, argv[1:])
		if cm == nil {
			return nil, false
		}
	*/
}

/*
// doRoute Recursive helper for Route
func doRoute(lf *leaf, argv []string) (*com.Command, []string) {
	if lf == nil {
		return nil, nil
	}
	val, found := lf.leaves[argv[0]]
	if found {
		ret, args := doRoute(val, argv[1:])
		if ret == nil {
			// Next node is end, use current lf and argv
			return lf.command, argv[1:]
		} else {
			// Next node isn't end, pass up
			return ret, args
		}
	} else {
		// End search, pass up argv
		return nil, argv
	}
}
*/

// init for discordgo things
func init() {
	token, exists := os.LookupEnv("TOKEN")
	if !exists {
		log.Fatal("Missing Discord API Key: Set env var $TOKEN")
	}

	var err error
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Println("Error:", err)
	}

	err = dg.Open()
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}
}

// command initialisation
func init() {
	ping := com.NewPing()
	for _, names := range ping.Names() {
		router.AddCommand(ping, strings.Split(names, " "))
	}
}

func main() {
	var err error
	dg.UpdateListeningStatus("you")

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		message := strings.TrimSpace(m.Content)

		if !strings.HasPrefix(message, prefix) {
			return
		}

		s.ChannelTyping(m.ChannelID)

		command, found := router.Route(message)
		if !found {
			s.ChannelMessageSend(m.ChannelID, "Error: Unknown command")
		}

		// Call handler
		str, err = command.MsgHandle()
		if err != nil {
			log.Println(err)
			s.ChannelMessageSend(m.ChannelID, err.String())
			return
		}
	})

	// Don't close the connection, wait for a kill signal
	log.Println("Logged in as: ", dg.State.User.ID)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	sig := <-sc
	log.Println("Received Signal: " + sig.String())
	log.Println("Bye!")
	dg.Close()
}
