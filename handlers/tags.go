package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	. "github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/utils"
)

/* In this file:

Tags 			- tags
TagsAdd 		- tags add
TagsRemove 		- tags remove
TagsGet 		- tags get
TagsList 		- tags list
TagsView 		- tags view
TagsPlatforms 	- tags platforms
TagsPing 		- tags ping
TagsPingMe 		- tags pingme

*/

const (
	TagsKey  = "fulltags"
	TEAL     = 0x008080
	TAG_LIM  = 32
	PLAT_LIM = 32
	USER_LIM = 32 // discord's nick limit is 32
)

var (
	ErrPlatTooLong = errors.New("your platform is too long, keep it under " + strconv.Itoa(PLAT_LIM) + " characters")
	ErrTagTooLong  = errors.New("your tag is too long, keep it under " + strconv.Itoa(TAG_LIM) + " characters")
	ErrNoTags      = errors.New("no tags found in database, add a tag to start it")
	ErrNoPlatform  = errors.New("no platform of that name, add a tag on that platform to create it")
	ErrNoUser      = errors.New("you don't have a tag on this platform, add one to the specified platform")
	ErrNoMention   = errors.New("use a proper @ping")
)

/* Storer: tags */

type tag struct {
	ID       string
	Tag      string
	Platform string
	PingMe   bool
}

type platform struct {
	Name  string
	Role  *discordgo.Role
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

func (t *Tags) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	// TODO: route help message magic etc.
	// If you're reading this as an example: This is a hack, do not do this.
	out := "Subcommands:"
	out += "\n" + utils.Code("add")
	out += "\n" + utils.Code("remove")
	out += "\n" + utils.Code("get")
	out += "\n" + utils.Code("list")
	out += "\n" + utils.Code("view")
	out += "\n" + utils.Code("platforms")
	out += "\n" + utils.Code("ping")
	out += "\n" + utils.Code("pingme")
	return NewSimpleSend(msg.ChannelID, out), nil
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

func (t *TagsAdd) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	var err error
	var tgs tags
	var out = NewSend(msg.ChannelID)

	if len(t.Tag) > 30 {
		return nil, ErrTagTooLong
	}

	if len(t.Platform) > 30 {
		return nil, ErrPlatTooLong
	}

	// get all tags
	err = DBGet(&tgs, TagsKey, &tgs)
	if err == ErrDBNotFound {
		tgs = tags{make(map[string]*platform)}
	} else if err != nil {
		return nil, err
	}

	// get platform
	var drl *discordgo.Role
	plt, ok := tgs.Platforms[t.Platform]
	if !ok {
		// create new role
		drl, err = ses.GuildRoleCreate(msg.GuildID)
		if err != nil {
			return nil, err
		}

		// edit the role
		ses.GuildRoleEdit(msg.GuildID, drl.ID, t.Platform, TEAL, false, drl.Permissions, true)

		// create new platform
		plt = &platform{
			Name:  t.Platform,
			Role:  drl,
			Users: make(map[string]*tag),
		}
		tgs.Platforms[t.Platform] = plt
		out.AddSimpleMessage("Creating new platform: " + utils.Code(t.Platform))
		// TODO: use reaction to verify
	}

	// set role, silently fails
	ses.GuildMemberRoleAdd(msg.GuildID, msg.Author.ID, plt.Role.ID)

	// add tag to platform
	plt.Users[msg.Author.ID] = &tag{
		ID:       msg.Author.ID,
		Tag:      t.Tag,
		Platform: t.Platform,
		PingMe:   true, // opt-out pings
	}

	// set tags
	_, _, err = DBSet(&tgs, TagsKey)
	if err != nil {
		return nil, err
	}

	out.AddSimpleMessage("Added tag " + utils.Code(t.Tag) + " for " + utils.Code(t.Platform))
	return out, nil
}

/* tags remove */

type TagsRemove struct {
	Platform string `arg:"platform"`
}

func NewTagsRemove() *TagsRemove { return &TagsRemove{} }

func (t *TagsRemove) Aliases() []string { return []string{"tags remove", "tags rm"} }

func (t *TagsRemove) Desc() string { return "Removes your tag from a platform" }

func (t *TagsRemove) Roles() []string { return nil }

func (t *TagsRemove) Chans() []string { return nil }

func (t *TagsRemove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	var err error
	var tgs tags
	var out = NewSend(msg.ChannelID)

	// get all tags
	err = DBGet(&tgs, TagsKey, &tgs)
	if err == ErrDBNotFound {
		return nil, ErrNoPlatform
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

	if utg.PingMe {
		// remove role, silently fails
		ses.GuildMemberRoleRemove(msg.GuildID, msg.Author.ID, plt.Role.ID)
	}

	// remove the tag
	delete(plt.Users, msg.Author.ID)
	out.AddSimpleMessage("Removed your tag from " + utils.Code(t.Platform))

	if len(plt.Users) == 0 {
		// remove the role from guild, silently fails
		ses.GuildRoleDelete(msg.GuildID, plt.Role.ID)

		// remove the platform
		delete(tgs.Platforms, t.Platform)
		out.AddSimpleMessage("Removing empty platform: " + utils.Code(t.Platform))
	}

	_, _, err = DBSet(&tgs, TagsKey)
	if err != nil {
		return nil, err
	}

	return out, nil
}

/* tags get */

type TagsGet struct {
	Platform string `arg:"platform"`
}

func NewTagsGet() *TagsGet { return &TagsGet{} }

func (t *TagsGet) Aliases() []string { return []string{"tags get"} }

func (t *TagsGet) Desc() string { return "Gets your tag from a platform" }

func (t *TagsGet) Roles() []string { return nil }

func (t *TagsGet) Chans() []string { return nil }

func (t *TagsGet) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	var err error
	var tgs tags

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

	utg, ok := plt.Users[msg.Author.ID]
	if !ok {
		return nil, ErrNoUser
	}

	return NewSimpleSend(msg.ChannelID, "Your tag is "+utils.Code(utg.Tag)+" for platform "+utils.Code(utg.Platform)), nil
}

/* tags list */

type TagsList struct {
	Platform string `arg:"platform"`
}

func NewTagsList() *TagsList { return &TagsList{} }

func (t *TagsList) Aliases() []string { return []string{"tags list", "tags ls", "tags view"} }

func (t *TagsList) Desc() string { return "Lists all tags on a platform" }

func (t *TagsList) Roles() []string { return nil }

func (t *TagsList) Chans() []string { return nil }

func (t *TagsList) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	var err error
	var tgs tags

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

	list := fmt.Sprintf(fmt.Sprintf("Ping? | %%-%ds | %%s\n", PLAT_LIM), "User", "Tag")
	for _, utg := range plt.Users {
		mem, err := ses.State.Member(msg.GuildID, msg.Author.ID)
		if err != nil {
			mem, err = ses.GuildMember(msg.GuildID, msg.Author.ID)
			if err != nil {
				return nil, err
			}
		}
		list += fmt.Sprintf(fmt.Sprintf("%%-%dt | %%-%ds | %%s\n", 5, USER_LIM),
			utg.PingMe,
			mem.Nick,
			utg.Tag)
	}

	return NewSimpleSend(msg.ChannelID, t.Platform+"'s tags:\n"+utils.Block(list)), nil
}

/* tags user */

type TagsUser struct {
	User string `arg:"user_id"`
}

func NewTagsUser() *TagsUser { return &TagsUser{} }

func (t *TagsUser) Aliases() []string { return []string{"tags user"} }

func (t *TagsUser) Desc() string { return "Lists all tags of a user" }

func (t *TagsUser) Roles() []string { return nil }

func (t *TagsUser) Chans() []string { return nil }

func (t *TagsUser) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	var err error
	var tgs tags
	var usr *discordgo.User

	// get user
	if len(msg.Mentions) > 0 {
		usr = msg.Mentions[0]
	} else {
		return nil, ErrNoMention
	}

	// get all tags
	err = DBGet(&tgs, TagsKey, &tgs)
	if err == ErrDBNotFound {
		return nil, ErrNoTags
	} else if err != nil {
		return nil, err
	}

	// collect user's tags
	var utgs []*tag
	for _, plt := range tgs.Platforms {
		utg, ok := plt.Users[usr.ID]
		if !ok {
			continue
		}
		utgs = append(utgs, utg)
	}

	list := fmt.Sprintf(fmt.Sprintf("Ping? | %%-%ds | %%s\n", USER_LIM), "User", "Tag")
	for _, utg := range utgs {
		list += fmt.Sprintf(fmt.Sprintf("%%-%dt | %%-%ds | %%s\n", 5, PLAT_LIM),
			utg.PingMe,
			utg.Platform,
			utg.Tag)
	}

	return NewSimpleSend(msg.ChannelID, usr.Username+"'s tags:\n"+utils.Block(list)), nil
}

/* tags platforms */

type TagsPlatforms struct{}

func NewTagsPlatforms() *TagsPlatforms { return &TagsPlatforms{} }

func (t *TagsPlatforms) Aliases() []string { return []string{"tags platforms"} }

func (t *TagsPlatforms) Desc() string { return "Lists all platforms" }

func (t *TagsPlatforms) Roles() []string { return nil }

func (t *TagsPlatforms) Chans() []string { return nil }

func (t *TagsPlatforms) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	var err error
	var tgs tags

	// get all tags
	err = DBGet(&tgs, TagsKey, &tgs)
	if err == ErrDBNotFound {
		return nil, ErrNoTags
	} else if err != nil {
		return nil, err
	}

	// iterate platforms
	list := "Platforms:\n"
	for _, plt := range tgs.Platforms {
		list += plt.Name + "  |  "
		list += strconv.Itoa(len(plt.Users)) + " tag(s)\n"
	}

	return NewSimpleSend(msg.ChannelID, list), nil
}

/* tags ping */

type TagsPing struct {
	Platform string   `arg:"platform"`
	Message  []string `arg:"message"`
}

func NewTagsPing() *TagsPing { return &TagsPing{} }

func (t *TagsPing) Aliases() []string { return []string{"tags ping", "ask", "ping tags"} }

func (t *TagsPing) Desc() string { return "Ping everyone with the given platform" }

func (t *TagsPing) Roles() []string { return nil }

func (t *TagsPing) Chans() []string { return nil }

func (t *TagsPing) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	var err error
	var tgs tags

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

	pings := utils.Bold(plt.Name)
	for _, utg := range plt.Users {
		var dusr *discordgo.User
		dusr, err = ses.User(utg.ID)
		if err != nil || !utg.PingMe {
			continue
		}
		pings += " " + dusr.Mention()
	}
	pings += " " + strings.Join(t.Message, " ")

	return NewSimpleSend(msg.ChannelID, pings), nil
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

func (t *TagsPingMe) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
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
	if t.PingMe {
		// set role, silently fails
		ses.GuildMemberRoleAdd(msg.GuildID, msg.Author.ID, plt.Role.ID)
	} else {
		// remove role, silently fails
		ses.GuildMemberRoleRemove(msg.GuildID, msg.Author.ID, plt.Role.ID)
	}
	_, _, err = DBSet(&tgs, TagsKey)
	if err != nil {
		return nil, err
	}

	out := "You now can"
	if !t.PingMe {
		out += "'t"
	}
	out += " be pinged for " + utils.Code(t.Platform)
	return NewSimpleSend(msg.ChannelID, out), nil
}

/* tags clean */

type TagsClean struct{}

func NewTagsClean() *TagsClean { return &TagsClean{} }

func (t *TagsClean) Aliases() []string { return []string{"tags clean"} }

func (t *TagsClean) Desc() string {
	return `Does a few things:
	- Cleans invalid tags from the entire tags database 
	- Creates the role for a platform if one does not exist
	- Double-checks that platform roles are assigned based on PingMe status`
}

func (t *TagsClean) Roles() []string { return nil }

func (t *TagsClean) Chans() []string { return nil }

func (t *TagsClean) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	var err error
	var tgs tags
	var out = NewSend(msg.ChannelID)

	// get all tags
	err = DBGet(&tgs, TagsKey, &tgs)
	if err == ErrDBNotFound {
		return nil, ErrNoPlatform
	} else if err != nil {
		return nil, err
	}

	// iterate platforms
	for pname, plt := range tgs.Platforms {
		// iterate users
		for uid, usr := range plt.Users {
		ROLE:
			// check user
			_, err = ses.GuildMember(msg.GuildID, uid)
			if err != nil {
				// couldn't find user, remove tag from db
				delete(plt.Users, uid)
				out.AddSimpleMessage("Removed user: " + uid)
				continue
			}

			// update roles, creates a new role if one does not exist for the platform
			if usr.PingMe {
				err = ses.GuildMemberRoleAdd(msg.GuildID, uid, plt.Role.ID)
			} else {
				err = ses.GuildMemberRoleRemove(msg.GuildID, uid, plt.Role.ID)
			}
			if err != nil {
				// ASSUME only error is role not existing in server
				// create new role
				var drl *discordgo.Role
				drl, err = ses.GuildRoleCreate(msg.GuildID)
				if err != nil {
					return nil, err
				}

				// edit the role
				ses.GuildRoleEdit(msg.GuildID, drl.ID, pname, TEAL, false, drl.Permissions, true)

				// add role to platform
				plt.Role = drl
				out.AddSimpleMessage("Re-created missing role for platform: " + utils.Code(pname))

				// retry
				goto ROLE
			}
		}

		// clean empty platforms
		if len(plt.Users) == 0 {
			// remove the role from guild, fails silently
			ses.GuildRoleDelete(msg.GuildID, plt.Role.ID)

			// remove the platform
			delete(tgs.Platforms, pname)
			out.AddSimpleMessage("Removed empty platform: " + utils.Code(pname))
		}
	}

	_, _, err = DBSet(&tgs, TagsKey)
	if err != nil {
		return nil, err
	}

	out.AddSimpleMessage("All Clean!")
	return out, nil
}
