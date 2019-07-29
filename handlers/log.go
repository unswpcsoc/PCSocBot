package handlers

import (
	"bytes"
	"log"
	"net/http"
	"regexp"
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
	delFunc func()
	filFunc func()

	msgCache = NewMapCache(CACHE_LIM)

	badWords = []*regexp.Regexp{
		regexp.MustCompile("(?i)kms"),
		regexp.MustCompile("(?i)kill[[:space:]]*myself"),
		regexp.MustCompile("(?i)kill[[:space:]]*me"),
	}
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
	order  []string
	cache  map[string]*discordgo.Message
	images map[string]*discordgo.File
}

// NewMapCache returns a new map cache
func NewMapCache(lim int) *MapCache {
	return &MapCache{
		limit:  lim,
		order:  make([]string, lim),
		cache:  make(map[string]*discordgo.Message),
		images: make(map[string]*discordgo.File),
	}
}

// Insert puts a key-value pair
func (m *MapCache) Insert(ky string, vl *discordgo.Message) {
	// just in case we have garbage keys...
	for len(m.order) >= m.limit {
		first := m.order[0]
		delete(m.cache, first)
		delete(m.images, first)
		m.order = m.order[1:len(m.order)]
	}
	m.cache[ky] = vl
	m.order = m.order[1:len(m.order)]
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

type Log struct {
	NilCommand
	Mode bool `arg:"mode"`
}

func NewLog() *Log { return &Log{} }

func (l *Log) Aliases() []string { return []string{"log"} }

func (l *Log) Desc() string {
	return "Moderation logging tool for deleted messages. This command controls all logging."
}

func (l *Log) Roles() []string { return []string{"mod"} }

func (l *Log) Subcommands() []Command { return []Command{&LogDelete{}, &LogFilter{}} }

func (l *Log) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	stat := ""
	if l.Mode {
		initDel(ses)
		initFill(ses)

		stat = "on"
	} else {
		delFunc()
		delFunc = nil
		filFunc()
		filFunc = nil

		stat = "off"
	}

	return NewSimpleSend(msg.ChannelID, "Logging has been turned "+stat), nil
}

type LogDelete struct {
	Log
	Mode bool `arg:"mode"`
}

/* log delete */

func NewLogDelete() *LogDelete { return &LogDelete{} }

func (l *LogDelete) Aliases() []string { return []string{"log delete", "log del"} }

func (l *LogDelete) Desc() string {
	return "This command controls logging of deleted messages."
}

func (l *LogDelete) Subcommands() []Command { return nil }

func (l *LogDelete) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	stat := ""
	if delFunc == nil {
		initDel(ses)
		stat = "on"
	} else {
		if delFunc != nil {
			delFunc()
			delFunc = nil
		}
		stat = "off"
	}

	return NewSimpleSend(msg.ChannelID, "MessageDelete logging has been turned "+stat), nil
}

/* log filter */

type LogFilter struct {
	Log
	Mode bool `arg:"mode"`
}

func NewLogFilter() *LogFilter { return &LogFilter{} }

func (l *LogFilter) Aliases() []string { return []string{"log filter", "log fil"} }

func (l *LogFilter) Desc() string {
	return "This command controls logging of messages containing bad words."
}

func (l *LogFilter) Subcommands() []Command { return nil }

func (l *LogFilter) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	stat := ""
	if filFunc == nil {
		initFill(ses)
		stat = "on"
	} else {
		if filFunc != nil {
			filFunc()
			filFunc = nil
		}
		stat = "off"
	}

	return NewSimpleSend(msg.ChannelID, "MessageFilter logging has been turned "+stat), nil
}

func initDel(ses *discordgo.Session) {
	tmp1 := ses.AddHandler(func(se *discordgo.Session, mc *discordgo.MessageCreate) {
		msg := mc.Message
		if msg.Author.ID == se.State.User.ID {
			return
		}
		msgCache.Insert(msg.ID+msg.ChannelID, msg)
	})

	tmp2 := ses.AddHandler(func(se *discordgo.Session, dm *discordgo.MessageDelete) {
		// get from cache
		dtd, img, ok := msgCache.Pop(dm.Message.ID + dm.Message.ChannelID)
		if !ok {
			log.Println("Warning: Cache miss on logged MessageDelete event.")
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

		// TODO: fix
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

		cha, err := ses.State.Channel(dtd.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}

		out.Embed = &discordgo.MessageEmbed{
			Title: "Deleted Message from " + cha.Name,
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
			out.File = img
		}

		se.ChannelMessageSendComplex(LOG_CHAN, out)
	})
	delFunc = func() {
		tmp1()
		tmp2()
	}
}

func initFill(ses *discordgo.Session) {
	filFunc = ses.AddHandler(func(se *discordgo.Session, mc *discordgo.MessageCreate) {
		msg := mc.Message
		if msg.Author.ID == se.State.User.ID {
			return
		}

		bad := false
		// check for BAD WORDS
		for _, bw := range badWords {
			if bw.MatchString(msg.Content) {
				// bad word detected
				bad = true
				break
			}
		}

		if !bad {
			return
		}

		// craft fields
		fields := []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Content:",
				Value:  msg.Content,
				Inline: false,
			},
		}

		se.ChannelMessageSendEmbed(LOG_CHAN, &discordgo.MessageEmbed{
			Title: "Bad Word Detected",
			Author: &discordgo.MessageEmbedAuthor{
				IconURL: msg.Author.AvatarURL(""),
				Name:    msg.Author.String(),
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: string(msg.Timestamp),
			},
			Fields: fields,
			Color:  EMB_COL,
		})
	})
}
