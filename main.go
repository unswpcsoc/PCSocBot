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
	comm "github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/router"
)

const (
	MESSAGE_LIMIT = 2000 // discord's message character limit
)

var (
	echo bool

	dgo    *discordgo.Session
	rtr    router.Router
	prefix = "!"

	errs = log.New(os.Stderr, "Error: ", log.Ltime)
)

// flag parsing
func init() {
	flag.BoolVar(&echo, "echo", false, "Enables echo mode")
	flag.Parse()
}

// discordgo things
func init() {
	key, exists := os.LookupEnv("KEY")
	if !exists {
		errs.Fatalln("Missing Discord API Key: Set env var $KEY")
	}

	var err error
	dgo, err = discordgo.New("Bot " + key)
	if err != nil {
		errs.Fatalln(err)
	}

	err = dgo.Open()
	if err != nil {
		errs.Fatalln(err)
	}
}

// command registration
func init() {
	rtr = router.NewRouter()
	rtr.Addcommand(comm.NewPing())
	rtr.AddCommand(comm.NewEcho())

	rtr.Addcommand(comm.NewRole("Weeb"))
	rtr.Addcommand(comm.NewRole("Meta"))
	rtr.Addcommand(comm.NewRole("Bookworm"))
}

func main() {
	defer dgo.Close()

	if echo {
		dgo.UpdateListeningStatus("stdin")
		dgo.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
			if m.Author.ID == s.State.User.ID {
				return
			}
			log.Printf("%s: %s\n", m.Author.Username, m.Content)
		})

		fmt.Print("Enter a channel ID: ")
		var outch string
		if _, err := fmt.Scanf("%s", &outch); err != nil {
			errs.Fatalln(err)
		}

		if _, err := dgo.Channel(outch); err != nil {
			errs.Fatalln(err)
		}

		outstr := ""
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\n>")

		got, err := reader.ReadString('\n')

		for err == nil {
			if len(outstr) > MESSAGE_LIMIT {
				fmt.Println("Error: Message above message limit")
				continue
			}

			_, err = dgo.ChannelMessageSend(outch, got)

			fmt.Print("\n>")
			got, err = reader.ReadString('\n')
		}

		return
	}

	dgo.UpdateListeningStatus("you")
	dgo.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		trm := strings.TrimSpace(m.Content)
		if !strings.HasPrefix(trm, prefix) || len(trm) == 1 {
			return
		}
		s.ChannelTyping(m.ChannelID)

		argv := strings.Split(trm[1:], " ")
		com, ind := rtr.Route(argv)
		if com == nil {
			// TODO help message routing
			s.ChannelMessageSend(m.ChannelID, "Error: Unknown command")
			return
		}
		snd, err := com.MsgHandle(s, m.Message, argv[ind:])
		if err != nil {
			errs.Println(err)
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		err = snd.Send(s)
		if err != nil {
			errs.Println(err)
		}
	})

	// Don't close the connection, wait for a kill signal
	log.Println("Logged in as:", dgo.State.User.ID)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	sig := <-sc

	log.Println("Received Signal: " + sig.String())
	log.Println("Bye!")
}
