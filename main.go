package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	dg *discordgo.Session
)

// init for discordgo things
func init() {
	token, exists := os.LookupEnv("TOKEN")
	if !exists {
		log.Fatal("Missing Discord API Key: TOKEN")
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

func main() {
	dg.UpdateListeningStatus("you")

	// Don't close the connection, wait for a kill signal
	log.Println("Logged in as: ", dg.State.User.ID)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	sig := <-sc
	log.Println("Received Signal: " + sig.String())
	log.Println("Bye!")
	dg.Close()
}
