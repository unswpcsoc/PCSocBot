package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	com "github.com/unswpcsoc/PCSocBot/commands"
)

const (
	MESSAGE_LIMIT = 2000
)

var (
	_testing = false

	echo   bool
	dg     *discordgo.Session
	router Router
	prefix = "!"
)

// leaf is a leaf of the command router tree.
type leaf struct {
	command com.Command
	leaves  map[string]*leaf
}

// NewLeaf returns a leaf with the command.
func NewLeaf(cm com.Command) *leaf {
	return &leaf{
		command: cm,
		leaves:  make(map[string]*leaf),
	}
}

// Router routes a command string to a command.
type Router struct {
	routes *leaf
}

// NewRouter returns a new Router structure.
func NewRouter() Router {
	return Router{NewLeaf(nil)}
}

// AddCommand adds command-string mapping
func (r *Router) AddCommand(cm com.Command, names []string) {
	// AddCommand(PingTagPeople, {"ask"})
	// AddCommand(PingTagPeople, {"tags", "ping"})
	if cm == nil || len(names) == 0 || r.routes == nil {
		return
	}

	for _, str := range names {
		argv := strings.Split(str, " ")

		// Search all known leaves
		curr := r.routes
		for {
			next, found := curr.leaves[argv[0]]
			if !found {
				break
			}
			curr = next
			argv = argv[1:]
		}

		// Add new leaves for remaining args
		for len(argv) > 0 {
			curr.leaves[argv[0]] = NewLeaf(nil)
			curr = curr.leaves[argv[0]]
			argv = argv[1:]
		}

		// Assign command to the final leaf
		curr.command = cm
	}
}

// Route routes to handler from string.
// Returns the command and the number of matched args.
// e.g.
//	   // r has a route through "example"->"command"->"string"
//     com, ind := r.Route([]string{"example", "command", "string", "with", "args"})
//
//	   // com will contain the command at "string" leaf
//	   // ind will be 3
func (r *Router) Route(argv []string) (com.Command, int) {
	if r.routes == nil || len(argv) == 0 {
		return nil, 0
	}

	// iterate through routes
	i := 0
	curr := r.routes
	var prev *leaf = nil
	var ok bool
	for i = 0; i < len(argv); i++ {
		curr, ok = curr.leaves[argv[i]]
		if !ok {
			break
		}
		prev = curr
	}

	return prev.command, i
}

// flag parsing
func init() {
	if _testing {
		return
	}

	flag.BoolVar(&echo, "echo", false, "Set this flag to enable echo mode")
	flag.Parse()
}

// discordgo things
func init() {
	log.Println("testing is:", _testing)
	if _testing {
		return
	}

	token, exists := os.LookupEnv("TOKEN")
	if !exists {
		log.Fatalln("Missing Discord API Key: Set env var $TOKEN")
	}

	var err error
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatalln("Error: ", err)
	}
}

// command registration
func init() {
	if _testing {
		return
	}

	router = NewRouter()

	ping := com.NewPing()
	router.AddCommand(ping, ping.Names())
}

func main() {
	defer dg.Close()
	dg.UpdateListeningStatus("you")

	if echo {
		fmt.Print("Enter a channel ID: ")

		var outch string
		if got, _ := fmt.Scanf("%s", &outch); got == 0 {
			log.Fatalln("Try again")
		}

		if _, err := dg.Channel(outch); err != nil {
			log.Fatalln(err)
		}

		var outstr string
		fmt.Print("\n>")
		got, _ := fmt.Scanln(&outstr)
		for got != 0 {
			if len(outstr) > MESSAGE_LIMIT {
				fmt.Println("Error: Message above message limit")
				continue
			}
			dg.ChannelMessageSend(outch, outstr)
			fmt.Print("\n>")
			got, _ = fmt.Scanln(&outstr)
		}
	} else {
		dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
			if m.Author.ID == s.State.User.ID {
				return
			}

			message := strings.TrimSpace(m.Content)

			if !strings.HasPrefix(message, prefix) {
				return
			}

			s.ChannelTyping(m.ChannelID)

			command, ind := router.Route(strings.Split(message, " "))
			if ind == 0 {
				// TODO: print help message
				s.ChannelMessageSend(m.ChannelID, "Error: Unknown command")
			}

			// Call handler
			send, err := command.MsgHandle(s, m.Message)
			if err != nil {
				log.Println(err)
				s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
				return
			}

			// SEND IT
			send.Send(s)
		})

		// Don't close the connection, wait for a kill signal
		log.Println("Logged in as: ", dg.State.User.ID)
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		sig := <-sc
		log.Println("Received Signal: " + sig.String())
	}

	log.Println("Bye!")
}
