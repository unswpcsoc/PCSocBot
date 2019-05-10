package main

import (
	"bufio"
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
	MESSAGE_LIMIT = 2000 // discord's message character limit
)

var (
	_testing = false

	echo    bool
	verbose bool

	dg     *discordgo.Session
	router Router
	prefix = "!"

	errs = log.New(os.Stderr)
)

// Leaf is a leaf of the command router tree.
type Leaf struct {
	command com.Command
	leaves  map[string]*Leaf
}

// NewLeaf returns a Leaf with the command.
func NewLeaf(cm com.Command) *leaf {
	return &Leaf{
		command: cm,
		leaves:  make(map[string]*Leaf),
	}
}

// Router routes a command string to a command.
type Router struct {
	routes *Leaf
}

// NewRouter returns a new Router structure.
func NewRouter() Router {
	return Router{NewLeaf(nil)}
}

// AddCommand adds command-string mapping
// e.g.
//     // c is of type Command
//     AddCommand(c, []string{"ask"})
//     AddCommand(c, []string{"tags", "ping"})
//
// Will register the commeand c to the command strings "ask" and "tags ping".
// Calling Route with these command strings will return the command mapped to it.
func (r *Router) AddCommand(cm com.Command, names []string) {
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
// `com` will contain the command at "string" leaf
// `ind` will be 3
func (r *Router) Route(argv []string) (com.Command, int) {
	if r.routes == nil || len(argv) == 0 {
		return nil, 0
	}

	// iterate through routes
	i := 0
	curr := r.routes
	var prev *Leaf = nil
	var ok bool
	for i = 0; i < len(argv); i++ {
		curr, ok = curr.leaves[argv[i]]
		if !ok {
			break
		}
		prev = curr
	}

	if prev == nil {
		return nil, i
	}

	return prev.command, i
}

// flag parsing
func init() {
	if _testing {
		return
	}

	flag.BoolVar(&echo, "echo", false, "Enables echo mode")
	flag.BoolVar(&verbose, "verbose", false, "Enables verbose mode")
	flag.Parse()
}

// discordgo things
func init() {
	if _testing {
		return
	}

	key, exists := os.LookupEnv("KEY")
	if !exists {
		errs.Fatalln("Missing Discord API Key: Set env var $KEY")
	}

	var err error
	dg, err = discordgo.New("Bot " + key)
	if err != nil {
		errs.Fatalln("Error:", err)
	}

	err = dg.Open()
	if err != nil {
		errs.Fatalln("Error: ", err)
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

	if echo {
		// echo mode
		dg.UpdateListeningStatus("stdin")

		fmt.Print("Enter a channel ID: ")
		var outch string
		if _, err := fmt.Scanf("%s", &outch); err != nil {
			errs.Fatalln(err)
		}

		if _, err := dg.Channel(outch); err != nil {
			errs.Fatalln(err)
		}

		outstr := ""
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\n>")

		got, err := reader.ReadString('\n')
		if verbose {
			log.Println("Got string:", got)
		}

		for err == nil {
			if len(outstr) > MESSAGE_LIMIT {
				fmt.Println("Error: Message above message limit")
				continue
			}

			msg, _ := dg.ChannelMessageSend(outch, got)
			if verbose {
				log.Println("Send message:", msg)
			}

			fmt.Print("\n>")
			got, err = reader.ReadString('\n')
			if verbose {
				log.Println("Got string:", got)
			}
		}

		return
	}
	dg.UpdateListeningStatus("you")

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		message := strings.TrimSpace(m.Content)
		if !strings.HasPrefix(message, prefix) || len(message) == 1 {
			return
		}

		s.ChannelTyping(m.ChannelID)

		cmdstr := message[1:]
		command, _ := router.Route(strings.Split(cmdstr, " "))
		if verbose {
			log.Println("Routed string", cmdstr, "to command", command)
		}
		if command == nil {
			// TODO help message routing
			s.ChannelMessageSend(m.ChannelID, "Error: Unknown command")
			return
		}

		// Call handler
		send, err := command.MsgHandle(s, m.Message)
		if verbose {
			log.Println("Handled message", m.Message, "\nHandler returned", send)
		}
		if err != nil {
			errs.Println(err)
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		// SEND IT
		send.Send(s)
	})

	// Don't close the connection, wait for a kill signal
	log.Println("Logged in as:", dg.State.User.ID)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	sig := <-sc

	log.Println("Received Signal: " + sig.String())
	log.Println("Bye!")
}
