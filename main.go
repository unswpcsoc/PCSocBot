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
	"github.com/unswpcsoc/PCSocBot/handlers"
	"github.com/unswpcsoc/PCSocBot/utils"
)

const (
	MESSAGE_LIMIT = 2000 // discord's message character limit
)

var (
	echo bool // echo mode
	prod bool // production mode i.e. db saves to file rather than memory
	sync bool // sync mode - will handle events syncronously if set

	lastCom = make(map[string]commands.Command) // map of uid->command for most recently used command

	dgo *discordgo.Session

	errs = log.New(os.Stderr, "Error: ", log.Ltime) // logger for errors
)

// flag parse init
func init() {
	flag.BoolVar(&echo, "echo", false, "Enables echo mode")
	flag.BoolVar(&prod, "prod", false, "Enables production mode")
	flag.BoolVar(&sync, "sync", false, "Enables synchronous event handling")
	flag.Parse()
}

// discordgo init
func init() {
	key, ok := os.LookupEnv("KEY")
	if !ok {
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

	dgo.SyncEvents = sync

	log.Printf("Logged in as: %v\nSyncEvents is %v", dgo.State.User.ID, dgo.SyncEvents)
}

// db init
func init() {
	var err error
	if prod {
		err = commands.DBOpen("./bot.db")
	} else {
		err = commands.DBOpen(":memory:")
	}
	if err != nil {
		errs.Fatalln(err)
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
		// catch panics on production
		if prod {
			defer func() {
				if r := recover(); r != nil {
					errs.Printf("Caught panic: %#v\n", r)
				}
			}()
		}

		if m.Author.ID == s.State.User.ID || m.Author.Bot {
			return
		}

		trm := strings.TrimSpace(m.Content)
		if !strings.HasPrefix(trm, commands.PREFIX) || len(trm) == 1 {
			return
		}
		s.ChannelTyping(m.ChannelID)

		// route message
		var com commands.Command
		var ind int
		var ok bool
		argv := strings.Split(trm[1:], " ")
		if argv[0] == "!" {
			com, ok = lastCom[m.Message.Author.ID]
			if !ok {
				out := utils.Italics("Error: You haven't issued any commands yet!")
				s.ChannelMessageSend(m.ChannelID, out)
				return
			}
			// !! args...
			ind = 1
		} else {
			// regular routing
			com, ind = handlers.RouterRoute(argv)
			if com == nil {
				out := utils.Italics("Error: Unknown command; use " + handlers.HELPALIAS)
				s.ChannelMessageSend(m.ChannelID, out)
				return
			}
		}

		// check chans
		chans := com.Chans()
		has, err := utils.MsgInChannels(s, m.Message, chans)
		if err != nil {
			errs.Printf("Channel checking threw: %#v\n", err)
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

		// check roles
		roles := com.Roles()
		has, err = utils.MsgHasRoles(s, m.Message, roles)
		if err != nil {
			errs.Printf("Role checking threw: %#v\n", err)
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

		// successfully routed, register in !! before usage check
		lastCom[m.Message.Author.ID] = com

		// fill args and check usage
		err = commands.FillArgs(com, argv[ind:])
		if err != nil {
			usage := "Usage: " + commands.GetUsage(com)
			s.ChannelMessageSend(m.ChannelID, usage)
			errs.Printf("Usage error on command %#v: %#v\n", com, err)
			return
		}

		// handle message
		snd, err := com.MsgHandle(s, m.Message)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, utils.Italics("Error: "+err.Error()))
			errs.Printf("%#v threw error: %#v\n", com, err)
			return
		}

		// send message
		err = snd.Send(s)
		if err != nil {
			errs.Printf("Send error: %#v\n", err)
		}

		// clean up args
		commands.CleanArgs(com)
	})

	// keep alive
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	sig := <-sc

	log.Println("Received Signal: " + sig.String())
	log.Println("Bye!")
}
