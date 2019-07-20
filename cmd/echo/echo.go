package main

import "fmt"

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

func main() {
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

	// keep alive
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	sig := <-sc

	log.Println("Received Signal: " + sig.String())
	log.Println("Bye!")
}
