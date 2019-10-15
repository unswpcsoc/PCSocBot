package handlers

import (
	"bytes"
	"errors"
	logs "log"
	"net/http"
	"regexp"
	//"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/commands"
)

const (
	cacheLimit  = 100
	embedColour = 0xff0000
	logChannel  = "529463078610534410" // #report
)

var (
	killDel func()
	killFil func()

	msgCache = NewMapCache(cacheLimit)

	badWords = []*regexp.Regexp{
		regexp.MustCompile("(?i)kms"),
		regexp.MustCompile("(?i)kill[[:space:]]*myself"),
		regexp.MustCompile("(?i)kill[[:space:]]*me"),
		regexp.MustCompile("(?i)retard"),
	}

	ErrLoggingOn  = errors.New("logging is already on")
	ErrLoggingOff = errors.New("logging is already off")
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
	// ensure order slice is below limit
	for len(m.order) >= m.limit {
		first := m.order[0]
		delete(m.cache, first)
		delete(m.images, first)
		m.order = m.order[1:len(m.order)]
	}
	// set kv
	m.cache[ky] = vl
	// append order
	m.order = append(m.order, ky)
	// try decode images and cache them too
	if len(vl.Attachments) > 0 {
		// get response, assuming only 1 attachment
		url := vl.Attachments[0].URL
		splits := strings.Split(url, ".")
		format := splits[len(splits)-1]
		logs.Println("Got attachment format: " + format)

		resp, err := http.Get(url)
		if err != nil {
			logs.Println(err)
			return
		}
		defer resp.Body.Close()

		// read into buffer
		var buf = bytes.NewBuffer([]byte{})
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			logs.Println(err)
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

type log struct {
	nilCommand
	Mode bool `arg:"mode"`
}

func newLog() *log { return &log{} }

func (l *log) Aliases() []string { return []string{"log"} }

func (l *log) Desc() string {
	return "Moderation logging tool for deleted messages. This command controls all logging."
}

func (l *log) Roles() []string { return []string{"mod"} }

func (l *log) Subcommands() []commands.Command {
	return []commands.Command{
		newLogDelete(),
		newLogFilter(),
	}
}

func (l *log) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	stat := ""
	if l.Mode {
		// TODO: test
		if killFil != nil || killDel != nil {
			return nil, ErrLoggingOn
		}

		initDel(ses)
		initFil(ses)

		stat = "on"
	} else {
		if killFil == nil && killDel == nil {
			return nil, ErrLoggingOff
		}

		killDel()
		killDel = nil
		killFil()
		killFil = nil

		stat = "off"
	}

	return commands.NewSimpleSend(msg.ChannelID, "logging has been turned "+stat), nil
}

type logDelete struct {
	log
	Mode bool `arg:"mode"`
}

func newLogDelete() *logDelete { return &logDelete{} }

func (l *logDelete) Aliases() []string { return []string{"log delete", "log del"} }

func (l *logDelete) Desc() string {
	return "This command controls logging of deleted messages."
}

func (l *logDelete) Subcommands() []commands.Command { return nil }

func (l *logDelete) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	stat := ""
	if l.Mode {
		if killDel != nil {
			return nil, ErrLoggingOn
		}

		initDel(ses)
		stat = "on"
	} else {
		if killDel == nil {
			return nil, ErrLoggingOff
		}

		killDel()
		killDel = nil
		stat = "off"
	}

	return commands.NewSimpleSend(msg.ChannelID, "MessageDelete logging has been turned "+stat), nil
}

type logFilter struct {
	log
	Mode bool `arg:"mode"`
}

func newLogFilter() *logFilter { return &logFilter{} }

func (l *logFilter) Aliases() []string { return []string{"log filter", "log fil"} }

func (l *logFilter) Desc() string {
	return "This command controls logging of messages containing bad words."
}

func (l *logFilter) Subcommands() []commands.Command { return nil }

func (l *logFilter) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	stat := ""
	if l.Mode {
		if killFil != nil {
			return nil, ErrLoggingOn
		}

		initFil(ses)
		stat = "on"
	} else {
		if killFil == nil {
			return nil, ErrLoggingOff
		}

		killFil()
		killFil = nil

		stat = "off"
	}

	return commands.NewSimpleSend(msg.ChannelID, "MessageFilter logging has been turned "+stat), nil
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
			logs.Println("Warning: Cache miss on logged MessageDelete event.")
			return
		}

		// craft message
		out := &discordgo.MessageSend{
			Content: "",
			Tts:     false,
		}

		// check if content was empty
		cnt := dtd.Content
		if len(cnt) == 0 {
			cnt = "[EMPTY]"
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
		/*
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
		*/

		cha, err := ses.State.Channel(dtd.ChannelID)
		if err != nil {
			logs.Println(err)
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
			Color:  embedColour,
		}

		if len(dtd.Attachments) > 0 {
			// assume only one attachment (limit for normal users)
			out.File = img
		}

		se.ChannelMessageSendComplex(logChannel, out)
	})
	killDel = func() {
		tmp1()
		tmp2()
	}
}

func initFil(ses *discordgo.Session) {
	killFil = ses.AddHandler(func(se *discordgo.Session, mc *discordgo.MessageCreate) {
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

		se.ChannelMessageSendEmbed(logChannel, &discordgo.MessageEmbed{
			Title: "Bad Word Detected",
			Author: &discordgo.MessageEmbedAuthor{
				IconURL: msg.Author.AvatarURL(""),
				Name:    msg.Author.String(),
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: string(msg.Timestamp),
			},
			Fields: fields,
			Color:  embedColour,
		})
	})
}
