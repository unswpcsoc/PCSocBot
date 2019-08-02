package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/1lann/staticice"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"

	"github.com/unswpcsoc/PCSocBot/commands"
)

const (
	letterBytes  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	staticFormat = "https://www.staticice.com.au/cgi-bin/search.cgi?price-min=%d&q=%s"
)

type staticIce struct {
	nilCommand
	Floor int      `arg:"price floor"`
	Query []string `arg:"search term"`
}

func newStaticIce() *staticIce { return &staticIce{} }

func (s *staticIce) Aliases() []string { return []string{"staticice", "static ice"} }

func (s *staticIce) Desc() string {
	return "Searches static ice and returns the top 10 results that are above the price floor"
}

func (s *staticIce) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error

	if s.Floor < 0 {
		return nil, errors.New("price floor can't be negative")
	}

	if len(s.Query) == 0 {
		return nil, errors.New("search term can't be empty")
	}

	// make query
	cli := staticice.NewClient(http.DefaultClient)
	res, err := cli.Search(
		staticice.RegionAU,
		staticice.NewSearchQuery().Query(strings.Join(s.Query, " ")).MinPrice(s.Floor),
	)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, errors.New("no search results for " + strings.Join(s.Query, " "))
	}

	if len(res) > 10 {
		res = res[0:10]
	}

	// get results
	var fields []*discordgo.MessageEmbedField
	for _, ent := range res {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("**%s**", ent.Seller),
			Value: fmt.Sprintf("**$%.2f**\nLink: %s\n%s\nLast Updated %s", ent.Price, ent.Link,
				ent.Description, humanize.Time(ent.LastUpdated)),
			Inline: false,
		})
	}

	// send it
	snd := commands.NewSend(msg.ChannelID).Embed(&discordgo.MessageEmbed{
		Title:  "Search Results",
		Color:  0x237AFC,
		Fields: fields,
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://ww.staticice.com.au",
			Name:    "StaticIce",
			IconURL: "https://www.staticice.com.au/images/logo.jpg",
		},
	})
	return snd, nil
}
