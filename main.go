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
	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/router"
	"github.com/unswpcsoc/PCSocBot/utils"
)

const (
	MESSAGE_LIMIT = 2000 // discord's message character limit
)

var (
	echo bool
	prod bool

	dgo    *discordgo.Session
	rtr    router.Router
	prefix = "!"

	errs = log.New(os.Stderr, "Error: ", log.Ltime)
)

// flag parsing
func init() {
	flag.BoolVar(&echo, "echo", false, "Enables echo mode")
	flag.BoolVar(&prod, "prod", false, "Enables production mode")
	flag.Parse()
}

// discordgo session
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

	rtr.Addcommand(commands.NewPing())

	rtr.Addcommand(commands.NewEcho())

	rtr.Addcommand(commands.NewQuote())
	rtr.Addcommand(commands.NewQuoteList())
	rtr.Addcommand(commands.NewQuotePending())
	rtr.Addcommand(commands.NewQuoteAdd())
	rtr.Addcommand(commands.NewQuoteApprove())
	rtr.Addcommand(commands.NewQuoteRemove())
	rtr.Addcommand(commands.NewQuoteReject())

	rtr.Addcommand(commands.NewDecimalSpiral())

	rtr.Addcommand(commands.NewRole("Weeb"))
	rtr.Addcommand(commands.NewRole("Meta"))
	rtr.Addcommand(commands.NewRole("Bookworm"))
}

// db intialisation
func init() {
	if prod {
		commands.DBOpen("./bot.db")
	} else {
		commands.DBOpen(":memory:")
	}
}

func main() {
	defer dgo.Close()
	defer commands.DBClose()

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

		// Route message
		argv := strings.Split(trm[1:], " ")
		com, ind := rtr.Route(argv)
		if com == nil {
			// TODO help message routing
			s.ChannelMessageSend(m.ChannelID, utils.Italics("Error: Unknown command"))
			return
		}

		// Check chans
		chans := com.Chans()
		has, err := utils.MsgInChannels(s, m.Message, chans)
		if err != nil {
			errs.Println(err)
		}
		if !has {
			out := "Error: You must be in " + utils.Code(chans[0])
			if len(chans) > 1 {
				others := chans[1:]
				for _, oth := range others {
					out += " or " + utils.Code(oth)
				}
			}
			out += " to use this command"
			s.ChannelMessageSend(m.ChannelID, utils.Italics(out))
			return
		}

		// Check roles
		roles := com.Roles()
		has, err = utils.MsgHasRoles(s, m.Message, roles)
		if err != nil {
			errs.Println(err)
		}
		if !has {
			out := "Error: You must be a " + utils.Code(roles[0])
			if len(roles) > 1 {
				others := roles[1:]
				for _, oth := range others {
					out += " or a " + utils.Code(oth)
				}
			}
			out += " to use this command"
			s.ChannelMessageSend(m.ChannelID, utils.Italics(out))
			return
		}

		// Handle message
		snd, err := com.MsgHandle(s, m.Message, argv[ind:])
		if err != nil {
			errs.Println(err)
			s.ChannelMessageSend(m.ChannelID, utils.Italics("Error: "+err.Error()))
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
