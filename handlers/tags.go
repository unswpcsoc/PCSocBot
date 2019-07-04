package handlers

import (
	"errors"

	"github.com/bwmarrin/discordgo"

	. "github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/utils"
)

const (
	TagsKey = "fulltags"
)

var (
	ErrNoTags     = errors.New("no tags found in database, add a tag to start it")
	ErrNoPlatform = errors.New("no platform of that name, add a tag on that platform to create it")
	ErrNoUser     = errors.New("you don't have a tag on this platform, add one to the specified platform")
)

/* In this file:

tags
tags add
tags remove
tags get
tags list
tags view
tags platforms
tags ping
tags pingme

*/

/* Storer: tags */

type tag struct {
	ID       string
	Tag      string
	Platform string
	PingMe   bool
}

type platform struct {
	Name string
	//TODO: Role string
	Users map[string]*tag // indexed by user id's
}

type tags struct {
	Platforms map[string]*platform
}

func (t *tags) Index() string { return "tags" }

/* tags */

type Tags struct{}

func NewTags() *Tags { return &Tags{} }

func (t *Tags) Aliases() []string { return []string{"tags"} }

func (t *Tags) Desc() string { return "Does nothing, literally just here for the help message." }

func (t *Tags) Roles() []string { return nil }

func (t *Tags) Chans() []string { return nil }

func (t *Tags) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	return NewSimpleSend(msg.ChannelID, "WIP"), nil
}

/* tags add */

type TagsAdd struct {
	Platform string `arg:"platform"`
	Tag      string `arg:"tag"`
	// TODO: default arg for PingMe
}

func NewTagsAdd() *TagsAdd { return &TagsAdd{} }

func (t *TagsAdd) Aliases() []string { return []string{"tags add"} }

func (t *TagsAdd) Desc() string { return "Adds your tag to a platform" }

func (t *TagsAdd) Roles() []string { return nil }

func (t *TagsAdd) Chans() []string { return nil }

func (t *TagsAdd) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	var err error
	var tgs tags
	var out = NewSend(msg.ChannelID)

	// get all tags
	err = DBGet(&tgs, TagsKey, &tgs)
	if err == ErrDBNotFound {
		tgs = tags{make(map[string]*platform)}
	} else if err != nil {
		return nil, err
	}

	// get platform
	plt, ok := tgs.Platforms[t.Platform]
	if !ok {
		// create new platform
		plt = &platform{
			Name:  t.Platform,
			Users: make(map[string]*tag),
		}
		tgs.Platforms[t.Platform] = plt
		out.AddSimpleMessage("Creating new platform: " + utils.Code(t.Platform))
		// TODO: use reaction to verify
	}

	// add tag to platform
	plt.Users[msg.Author.ID] = &tag{
		ID:       msg.Author.ID,
		Tag:      t.Tag,
		Platform: t.Platform,
		PingMe:   true, // TODO: decide opt-in or opt-out
	}

	// set tags
	_, _, err = DBSet(&tgs, TagsKey)
	if err != nil {
		return nil, err
	}

	out.AddSimpleMessage("Added tag " + utils.Code(t.Tag) + " for platform " + utils.Code(t.Platform))
	return out, nil
}

/* tags remove */

type TagsRemove struct{}

func NewTagsRemove() *TagsRemove { return &TagsRemove{} }

func (t *TagsRemove) Aliases() []string { return []string{"tags remove", "tags rm"} }

func (t *TagsRemove) Desc() string { return "Removes your tag from a platform" }

func (t *TagsRemove) Roles() []string { return nil }

func (t *TagsRemove) Chans() []string { return nil }

func (t *TagsRemove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	return nil, nil
}

/* tags get */

type TagsGet struct{}

func NewTagsGet() *TagsGet { return &TagsGet{} }

func (t *TagsGet) Aliases() []string { return []string{"tags get"} }

func (t *TagsGet) Desc() string { return "Gets your tag from a platform" }

func (t *TagsGet) Roles() []string { return nil }

func (t *TagsGet) Chans() []string { return nil }

func (t *TagsGet) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	return nil, nil
}

/* tags list */

type TagsList struct {
	Platform string `arg:"platform"`
}

func NewTagsList() *TagsList { return &TagsList{} }

func (t *TagsList) Aliases() []string { return []string{"tags list", "tags ls"} }

func (t *TagsList) Desc() string { return "Lists all tags on a platform" }

func (t *TagsList) Roles() []string { return nil }

func (t *TagsList) Chans() []string { return nil }

func (t *TagsList) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	var err error
	var tgs tags
	// get all tags
	err = DBGet(&tgs, TagsKey, &tgs)
	if err == ErrDBNotFound {
		return nil, ErrNoTags
	} else if err != nil {
		return nil, err
	}

	plt, ok := tgs.Platforms[t.Platform]
	if !ok {
		return nil, ErrNoPlatform
	}

	list := utils.Code(t.Platform) + "'s tags:\n"
	// TODO: construct embed for this
	for _, utg := range plt.Users {
		var dusr *discordgo.User
		dusr, err = ses.User(utg.ID)
		if err != nil {
			// silently fail
			continue
		}
		list += dusr.Username + " : "
		list += utg.Tag + " : can"
		if utg.PingMe {
			list += "'t"
		}
		list += " ping\n"
	}

	return NewSimpleSend(msg.ChannelID, list), nil
}

/* tags view */

type TagsView struct{}

func NewTagsView() *TagsView { return &TagsView{} }

func (t *TagsView) Aliases() []string { return []string{"tags view"} }

func (t *TagsView) Desc() string { return "Lists all tags of a user" }

func (t *TagsView) Roles() []string { return nil }

func (t *TagsView) Chans() []string { return nil }

func (t *TagsView) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	var err error
	var tgs tags
	// get all tags
	err = DBGet(&tgs, TagsKey, &tgs)
	if err == ErrDBNotFound {
		return nil, ErrNoTags
	} else if err != nil {
		return nil, err
	}

	// collect user tags
	var utgs []*tag
	for _, plt := range tgs.Platforms {
		utg, ok := plt.Users[msg.Author.ID]
		if !ok {
			continue
		}
		utgs = append(utgs, utg)
	}

	//

	list := "Your tags:\n"
	// TODO: construct embed for this
	for _, utg := range utgs {
		list += utg.Platform + "  |  "
		list += utg.Tag + "  |  "
		if utg.PingMe {
			list += " no"
		}
		list += " ping\n"
	}

	return NewSimpleSend(msg.ChannelID, list), nil
}

/* tags platforms */

type TagsPlatforms struct{}

func NewTagsPlatforms() *TagsPlatforms { return &TagsPlatforms{} }

func (t *TagsPlatforms) Aliases() []string { return []string{"tags platforms"} }

func (t *TagsPlatforms) Desc() string { return "Lists all platforms" }

func (t *TagsPlatforms) Roles() []string { return nil }

func (t *TagsPlatforms) Chans() []string { return nil }

func (t *TagsPlatforms) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	return nil, nil
}

/* tags ping */

type TagsPing struct{}

func NewTagsPing() *TagsPing { return &TagsPing{} }

func (t *TagsPing) Aliases() []string { return []string{"tags ping", "ask", "ping tags"} }

func (t *TagsPing) Desc() string { return "Ping everyone with the given platform" }

func (t *TagsPing) Roles() []string { return nil }

func (t *TagsPing) Chans() []string { return nil }

func (t *TagsPing) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	return nil, nil
}

/* tags pingme */

type TagsPingMe struct {
	Platform string `arg:"platform"`
	PingMe   bool   `arg:"wants pings"`
}

func NewTagsPingMe() *TagsPingMe { return &TagsPingMe{} }

func (t *TagsPingMe) Aliases() []string { return []string{"tags pingme", "askme"} }

func (t *TagsPingMe) Desc() string { return "Set your ping status for a given platform" }

func (t *TagsPingMe) Roles() []string { return nil }

func (t *TagsPingMe) Chans() []string { return nil }

func (t *TagsPingMe) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	var err error
	var tgs tags

	// get all tags
	err = DBGet(&tags{}, TagsKey, &tgs)
	if err == ErrDBNotFound {
		return nil, ErrNoTags
	} else if err != nil {
		return nil, err
	}

	// get platform
	plt, ok := tgs.Platforms[t.Platform]
	if !ok {
		return nil, ErrNoPlatform
	}

	// get tag
	utg, ok := plt.Users[msg.Author.ID]
	if !ok {
		return nil, ErrNoUser
	}

	// set pingme
	utg.PingMe = t.PingMe
	_, _, err = DBSet(&tgs, TagsKey)
	if err != nil {
		return nil, err
	}

	out := "You can"
	if !t.PingMe {
		out += "'t"
	}
	out += " be pinged for platform " + utils.Code(t.Platform)
	return NewSimpleSend(msg.ChannelID, out), nil
}
