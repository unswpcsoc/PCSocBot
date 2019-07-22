package handlers

import (
	"bytes"
	"net/http"

	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	. "github.com/unswpcsoc/PCSocBot/commands"
)

const (
	CACHE_LIM = 100
	EMB_COL   = 0xff0000
	LOG_CHAN  = "462063414408249376" // currently just the #commands channel in pcsoc2
)

var (
	msgFunc  func()
	delFunc  func()
	msgCache = NewMapCache(CACHE_LIM)
)

/* helpers */

// MapCache is a data structure that performs FIFOs insertions to stay under the limit,
// with indexing support
//
// This implementation uses discordgo messages but can be made generic with interface{}.
// I just picked a concrete type to avoid typing out a lot of type assertions.
//
// Also need to cache images as well since deleted messages remove all attachments
type MapCache struct {
	limit  int
	order  chan string
	cache  map[string]*discordgo.Message
	images map[string]*discordgo.File
}

// NewMapCache returns a new map cache
func NewMapCache(lim int) *MapCache {
	return &MapCache{
		limit:  lim,
		order:  make(chan string, lim),
		cache:  make(map[string]*discordgo.Message),
		images: make(map[string]*discordgo.File),
	}
}

// Insert puts a key-value pair
func (m *MapCache) Insert(ky string, vl *discordgo.Message) {
	// just in case we have garbage keys...
	for len(m.order) >= m.limit {
		first := <-m.order
		delete(m.cache, first)
		delete(m.images, first)
	}
	m.cache[ky] = vl
	m.order <- ky
	// try decode images and cache them too
	if len(vl.Attachments) > 0 {
		// get response, assuming only 1 attachment
		url := vl.Attachments[0].URL
		splits := strings.Split(url, ".")
		format := splits[len(splits)-1]
		log.Println("Got attachment format: " + format)

		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()

		// read into buffer
		var buf = bytes.NewBuffer([]byte{})
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}

		// chuck buffer into images
		m.images[ky] = &discordgo.File{
			Name:        "deleted attachment." + format,
			ContentType: "image/" + format,
			Reader:      buf,
		}
	}
}

// Pop gets a value given a key and removes it from the map
func (m *MapCache) Pop(ky string) (*discordgo.Message, *discordgo.File, bool) {
	vl, ok := m.cache[ky]
	if !ok {
		return nil, nil, false
	}
	im, _ := m.images[ky]
	delete(m.cache, ky)
	delete(m.images, ky)
	return vl, im, true
}

/* log */

type Log struct{ NilCommand }

func NewLog() *Log { return &Log{} }

func (l *Log) Aliases() []string { return []string{"log"} }

func (l *Log) Desc() string {
	return "Moderation logging tool for deleted messages. This command toggles logging."
}

func (l *Log) Roles() []string { return []string{"mod"} }

func (l *Log) Chans() []string { return []string{"mods"} }

func (l *Log) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	stat := ""
	if msgFunc == nil {
		msgFunc = ses.AddHandler(func(se *discordgo.Session, mc *discordgo.MessageCreate) {
			msg := mc.Message
			if msg.Author.ID == se.State.User.ID {
				return
			}
			msgCache.Insert(msg.ID+msg.ChannelID, msg)
		})

		delFunc = ses.AddHandler(func(se *discordgo.Session, dm *discordgo.MessageDelete) {
			// get from cache
			dtd, img, ok := msgCache.Pop(dm.Message.ID + dm.Message.ChannelID)
			if !ok {
				log.Println("Warning: MessageDelete event on uncached message.")
				return
			}

			// craft message
			out := &discordgo.MessageSend{
				Content: "",
				Tts:     false,
			}

			// craft fields
			fields := []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Content:",
					Value:  dtd.Content,
					Inline: false,
				},
			}

			if len(dtd.Reactions) > 0 {
				var reacts string
				for _, react := range dtd.Reactions {
					reacts += "[" + strconv.Itoa(react.Count) + "] "
					reacts += react.Emoji.Name
				}
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   "Reactions:",
					Value:  reacts,
					Inline: false,
				})
			}

			out.Embed = &discordgo.MessageEmbed{
				Title: "Deleted Message",
				Author: &discordgo.MessageEmbedAuthor{
					IconURL: dtd.Author.AvatarURL(""),
					Name:    dtd.Author.String(),
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: string(dtd.Timestamp),
				},
				Fields: fields,
				Color:  EMB_COL,
			}

			if len(dtd.Attachments) > 0 {
				// assume only one attachment (limit for normal users)
				out.Files = []*discordgo.File{img}
			}

			se.ChannelMessageSendComplex(LOG_CHAN, out)
		})
		stat = "on"
	} else {
		msgFunc()
		delFunc()
		stat = "off"
		msgFunc = nil
		delFunc = nil
	}

	return NewSimpleSend(msg.ChannelID, "Logging is now "+stat), nil
}
