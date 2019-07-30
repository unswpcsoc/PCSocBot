package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/utils"
)

const (
	emojiConfirm     = string(0x2714)
	emojiDeny        = string(0x274C)
	guildMemberLimit = 1000
	selfUserString   = "$ME"
	tagsKey          = "fulltags"
	teal             = 0x008080

	tagLimit  = 32
	platLimit = 32
	userLimit = 32 // discord's nick limit is 32
)

var (
	// ErrPlatTooLong means the user tried to create a platform that was too damn long
	ErrPlatTooLong = errors.New("your platform is too long, keep it under " + strconv.Itoa(platLimit) + " characters")
	// ErrTagTooLong means the user tried to create a tag that was too damn long
	ErrTagTooLong = errors.New("your tag is too long, keep it under " + strconv.Itoa(tagLimit) + " characters")
	// ErrNoTags means there is no tag list
	ErrNoTags = errors.New("no tags found in database, add a tag to start it")
	// ErrNoUserTags means there are no tags for the queried user
	ErrNoUserTags = errors.New("no tags found for that user")
	// ErrNoPlatform means the user queried a platform that doesn't exist
	ErrNoPlatform = errors.New("no platform of that name, add a tag on that platform to create it")
	// ErrNoUser means the user queried a platform they did not have a tag on
	ErrNoUser = errors.New("you don't have a tag on this platform, add one to the specified platform")
	// ErrUserNotFound means the user queried a username that doesn't exist on the server
	ErrUserNotFound = errors.New("user not found")
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

type tagStorer struct {
	Platforms map[string]*platform
}

func (t *tagStorer) Index() string { return "tags" }

/* tags */

type tags struct {
	nilCommand
}

func newTags() *tags { return &tags{} }

func (t *tags) Aliases() []string { return []string{"tags"} }

func (t *tags) Desc() string { return "tags root command." }

func (t *tags) Subcommands() []commands.Command {
	return []commands.Command{
		newTagsAdd(),
		newTagsClean(),
		newTagsGet(),
		newTagsList(),
		newTagsPlatforms(),
		newTagsPing(),
		newTagsPingMe(),
		newTagsRemove(),
		newTagsUser(),
	}
}

func (t *tags) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	return commands.NewSimpleSend(msg.ChannelID, commands.GetUsage(t)), nil
}

/* tags add */

type tagsAdd struct {
	nilCommand
	Platform string `arg:"platform"`
	Tag      string `arg:"tag"`
}

func newTagsAdd() *tagsAdd { return &tagsAdd{} }

func (t *tagsAdd) Aliases() []string { return []string{"tags add"} }

func (t *tagsAdd) Desc() string { return "Adds your tag to a platform" }

func (t *tagsAdd) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer
	var out = commands.NewSend(msg.ChannelID)

	if len(t.Tag) > 30 {
		return nil, ErrTagTooLong
	}

	if len(t.Platform) > 30 {
		return nil, ErrPlatTooLong
	}

	// get all tags
	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
		tgs = tagStorer{make(map[string]*platform)}
	} else if err != nil {
		return nil, err
	}

	// get platform
	var drl *discordgo.Role
	plt, ok := tgs.Platforms[t.Platform]
	if !ok {
		// wait for user reaction to verify
		war, _ := ses.ChannelMessageSend(msg.ChannelID, "Warning: Creating a new platform. Make sure you check if a similar one exists.\n"+
			"React to confirm this action.")

		// react to the message to get things going
		err = ses.MessageReactionAdd(war.ChannelID, war.ID, emojiConfirm)
		if err != nil {
			return nil, err
		}

		err = ses.MessageReactionAdd(war.ChannelID, war.ID, emojiDeny)
		if err != nil {
			return nil, err
		}

		// spin up goroutine to check if message has been reacted
		reaction := make(chan bool)
		go func() {
			reacted := make(chan int)
			kill := ses.AddHandler(func(se *discordgo.Session, no *discordgo.MessageReactionAdd) {
				// demo sonnanja dame
				// mou sonnanja hora
				// KOKORO WA SHINKA SURU YO
				// MOTTO
				// MOTTO

				// make sure reaction is on the correct message by the correct user
				if no.MessageReaction.MessageID != war.ID || no.MessageReaction.UserID != msg.Author.ID {
					return
				}

				// signal that we have achieved nirvana
				switch no.MessageReaction.Emoji.Name {
				case emojiConfirm:
					reaction <- true
					reacted <- 0
				case emojiDeny:
					reaction <- false
					reacted <- 0
				}
			})
			<-reacted
			kill()
		}()

		// check what we got
		if !<-reaction {
			return commands.NewSimpleSend(msg.ChannelID, "Aborting platform creation."), nil
		}

		// create new role
		drl, err = ses.GuildRoleCreate(msg.GuildID)
		if err != nil {
			return nil, err
		}

		// edit the role
		ses.GuildRoleEdit(msg.GuildID, drl.ID, t.Platform, teal, false, drl.Permissions, true)

		// create new platform
		plt = &platform{
			Name:  t.Platform,
			Role:  drl,
			Users: make(map[string]*tag),
		}
		tgs.Platforms[t.Platform] = plt
		out.AddSimpleMessage("Creating new platform: " + utils.Code(t.Platform))
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
	_, _, err = commands.DBSet(&tgs, tagsKey)
	if err != nil {
		return nil, err
	}

	out.AddSimpleMessage("Added tag " + utils.Code(t.Tag) + " for " + utils.Code(t.Platform))
	return out, nil
}

/* tags clean */

type tagsClean struct {
	nilCommand
}

func newTagsClean() *tagsClean { return &tagsClean{} }

func (t *tagsClean) Aliases() []string { return []string{"tags clean"} }

func (t *tagsClean) Desc() string {
	return `Does a few things:
	- Cleans invalid tags from the entire tags database 
	- Creates the role for a platform if one does not exist
	- Double-checks that platform roles are assigned based on PingMe status`
}

func (t *tagsClean) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer
	var out = commands.NewSend(msg.ChannelID)

	// get all tags
	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
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
				ses.GuildRoleEdit(msg.GuildID, drl.ID, pname, teal, false, drl.Permissions, true)

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

	_, _, err = commands.DBSet(&tgs, tagsKey)
	if err != nil {
		return nil, err
	}

	out.AddSimpleMessage("All Clean!")
	return out, nil
}

/* tags get */

type tagsGet struct {
	nilCommand
	Platform string `arg:"platform"`
}

func newTagsGet() *tagsGet { return &tagsGet{} }

func (t *tagsGet) Aliases() []string { return []string{"tags get"} }

func (t *tagsGet) Desc() string { return "Gets your tag for a platform." }

func (t *tagsGet) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer

	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
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

	return commands.NewSimpleSend(msg.ChannelID, "Your tag is "+utils.Code(utg.Tag)+" for platform "+utils.Code(utg.Platform)), nil
}

/* tags list */

type tagsList struct {
	nilCommand
	Platform string `arg:"platform"`
}

func newTagsList() *tagsList { return &tagsList{} }

func (t *tagsList) Aliases() []string { return []string{"tags list", "tags ls"} }

func (t *tagsList) Desc() string { return "Lists all tags for that platform." }

func (t *tagsList) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer

	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
		return nil, ErrNoTags
	} else if err != nil {
		return nil, err
	}

	plt, ok := tgs.Platforms[t.Platform]
	if !ok {
		return nil, ErrNoPlatform
	}

	list := fmt.Sprintf(fmt.Sprintf("Ping? | %%-%ds | %%s\n", platLimit), "User", "Tag")
	for _, utg := range plt.Users {
		mem, err := ses.State.Member(msg.GuildID, msg.Author.ID)
		if err != nil {
			mem, err = ses.GuildMember(msg.GuildID, msg.Author.ID)
			if err != nil {
				return nil, err
			}
		}
		list += fmt.Sprintf(fmt.Sprintf("%%-%dt | %%-%ds | %%s\n", 5, userLimit),
			utg.PingMe,
			mem.Nick,
			utg.Tag)
	}

	return commands.NewSimpleSend(msg.ChannelID, t.Platform+"'s tags:\n"+utils.Block(list)), nil
}

/* tags platforms */

type tagsPlatforms struct {
	nilCommand
}

func newTagsPlatforms() *tagsPlatforms { return &tagsPlatforms{} }

func (t *tagsPlatforms) Aliases() []string { return []string{"tags platforms"} }

func (t *tagsPlatforms) Desc() string { return "Lists all platforms." }

func (t *tagsPlatforms) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer

	// get all tags
	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
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

	return commands.NewSimpleSend(msg.ChannelID, list), nil
}

/* tags ping */

type tagsPing struct {
	nilCommand
	Platform string   `arg:"platform"`
	Message  []string `arg:"message"`
}

func newTagsPing() *tagsPing { return &tagsPing{} }

func (t *tagsPing) Aliases() []string { return []string{"tags ping", "ask", "ping tags"} }

func (t *tagsPing) Desc() string {
	return "Pings all users with `PingMe` set on the platform. Can also add your own message."
}

func (t *tagsPing) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer

	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
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

	return commands.NewSimpleSend(msg.ChannelID, pings), nil
}

/* tags pingme */

type tagsPingMe struct {
	nilCommand
	Platform string `arg:"platform"`
	PingMe   bool   `arg:"wants pings"`
}

func newTagsPingMe() *tagsPingMe { return &tagsPingMe{} }

func (t *tagsPingMe) Aliases() []string { return []string{"tags pingme", "askme"} }

func (t *tagsPingMe) Desc() string { return "Set your ping status for a given platform" }

func (t *tagsPingMe) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer

	// get all tags
	err = commands.DBGet(&tagStorer{}, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
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
	_, _, err = commands.DBSet(&tgs, tagsKey)
	if err != nil {
		return nil, err
	}

	out := "You now can"
	if !t.PingMe {
		out += "'t"
	}
	out += " be pinged for " + utils.Code(t.Platform)
	return commands.NewSimpleSend(msg.ChannelID, out), nil
}

/* tags remove */

type tagsRemove struct {
	nilCommand
	Platform string `arg:"platform"`
}

func newTagsRemove() *tagsRemove { return &tagsRemove{} }

func (t *tagsRemove) Aliases() []string { return []string{"tags remove", "tags rm"} }

func (t *tagsRemove) Desc() string { return "Removes your tag from a platform" }

func (t *tagsRemove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer
	var out = commands.NewSend(msg.ChannelID)

	// get all tags
	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
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

	_, _, err = commands.DBSet(&tgs, tagsKey)
	if err != nil {
		return nil, err
	}

	return out, nil
}

/* tags user */

type tagsUser struct {
	nilCommand
	User string `arg:"user_id"`
}

func newTagsUser() *tagsUser { return &tagsUser{} }

func (t *tagsUser) Aliases() []string { return []string{"tags user", "tags view"} }

func (t *tagsUser) Desc() string { return "Lists all tags of a user" }

func (t *tagsUser) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer
	var usr *discordgo.User

	// get user
	if t.User == selfUserString {
		usr = msg.Author
	} else {
		members, err := ses.GuildMembers(msg.GuildID, "0", guildMemberLimit)
		if err != nil {
			return nil, err
		}

		for _, mem := range members {
			if mem.User.Username == t.User {
				usr = mem.User
			}
		}

		if usr == nil {
			return nil, ErrUserNotFound
		}
	}

	// get all tags
	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
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

	if len(utgs) == 0 {
		return nil, ErrNoUserTags
	}

	list := fmt.Sprintf(fmt.Sprintf("Ping? | %%-%ds | %%s\n", userLimit), "User", "Tag")
	for _, utg := range utgs {
		list += fmt.Sprintf(fmt.Sprintf("%%-%dt | %%-%ds | %%s\n", 5, platLimit),
			utg.PingMe,
			utg.Platform,
			utg.Tag)
	}

	return commands.NewSimpleSend(msg.ChannelID, usr.Username+"'s tags:\n"+utils.Block(list)), nil
}
