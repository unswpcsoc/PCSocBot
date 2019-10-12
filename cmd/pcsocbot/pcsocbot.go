// This package contains an implementation of PCSocBot
// using the provided utilities.
//
// You can make your own if you like I guess.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	//"github.com/sahilm/fuzzy"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/handlers"
	"github.com/unswpcsoc/PCSocBot/internal/utils"
)

var (
	prod bool // production mode i.e. db saves to file rather than memory
	sync bool // sync mode - will handle events syncronously if set, might break things if you do this

	lastCom = make(map[string]commands.Command) // map of uid->command for most recently used command

	dgo *discordgo.Session

	errs = log.New(os.Stderr, "Error: ", log.Ltime) // logger for errors
)

// flag parse init
func init() {
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

	dgo.UpdateStatus(0, commands.Prefix+handlers.HelpAlias)

	// init loggers
	handlers.InitLogs(dgo)

	// handle commands
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
		if !strings.HasPrefix(trm, commands.Prefix) || len(trm) == 1 {
			return
		}

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
				/* REDACTED
				mat := fuzzy.Find(strings.Join(argv, " "), handlers.RouterToStringSlice())

				var out string
				if len(mat) > 0 {
					out = utils.Italics("Error: Unknown command, did you mean:") + "\n"

					// fuzzy find suggestions
					for i, m := range mat {
						// only get 3 suggestions at most
						if i == 3 {
							break
						}
						out += utils.Code(commands.Prefix+m.Str) + "\n"
					}
				} else {
					out = utils.Italics("Error: Unknown command") + "\n"
				}

				out += utils.Italics("Use") + " " + commands.Prefix + utils.Code(handlers.HelpAlias) + " " + utils.Italics("for more.")
				s.ChannelMessageSend(m.ChannelID, out)
				*/
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
		s.ChannelTyping(m.ChannelID)
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
