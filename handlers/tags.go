package handlers

import (
	"errors"
	"fmt"
	"golang.org/x/sync/semaphore"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/internal/utils"
)

const (
	emojiConfirm     = string(0x2705)
	emojiClean       = string(0x2728)
	emojiDeny        = string(0x274C)
	guildMemberLimit = 1000
	tagsKey          = "fulltags"
	teal             = 0x008080

	tagLimit  = 32
	platLimit = 20
	userLimit = 20 // discord's nick limit is 32

	addTimeout = 7
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
	ErrNoPlatform = errors.New("no platform of that name")
	// ErrNoUser means the user queried a platform they did not have a tag on
	ErrNoUser = errors.New("you don't have a tag on this platform")
	// ErrUserNotFound means the user queried a username that doesn't exist on the server
	ErrUserNotFound = errors.New("user not found")
	// ErrAddSpam means the user tried to add while a new platform was being waited on
	ErrAddSpam = errors.New("please do not try add anything while I'm waiting")
	// ErrCleanSpam means the user tried to clean while a clean is in progress
	ErrCleanSpam = errors.New("already cleaning, please be patient")

	// syncs
	addSemaphore   = semaphore.NewWeighted(1)
	cleanSemaphore = semaphore.NewWeighted(1)
)

type tag struct {
	UID      string
	Username string // don't trust this, always fetch from the UID
	Tag      string
	Platform string
	PingMe   bool
}

type platform struct {
	Name  string
	Role  *discordgo.Role
	Users map[string]*tag // indexed by user id's
}

// TODO: default games and api integrations
type tagStorer struct {
	Platforms map[string]*platform
}

func (t *tagStorer) Index() string { return "tags" }

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
		newTagsModRemove(),
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

type tagsAdd struct {
	nilCommand
	Platform string   `arg:"platform"`
	Tag      []string `arg:"tag"`
}

func newTagsAdd() *tagsAdd { return &tagsAdd{} }

func (t *tagsAdd) Aliases() []string { return []string{"tags add", "tags edit"} }

func (t *tagsAdd) Desc() string { return "Adds your tag to a platform" }

func (t *tagsAdd) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer
	var out = commands.NewSend(msg.ChannelID)

	if len(t.Tag) == 0 {
		return nil, errors.New("please provide a tag")
	}
	argTag := strings.Join(t.Tag, " ")
	if len(argTag) > 30 {
		return nil, ErrTagTooLong
	}

	if len(t.Platform) > 30 {
		return nil, ErrPlatTooLong
	}

	// check if we're already adding
	if !addSemaphore.TryAcquire(1) {
		return out.Message("Please don't add anything while I'm waiting!"), nil
	}
	defer addSemaphore.Release(1)

	// lock the db
	commands.DBLock()
	defer commands.DBUnlock()

	// get all tags
	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
		tgs = tagStorer{make(map[string]*platform)}
	} else if err != nil {
		return nil, err
	}

	// get platform
	//var drl *discordgo.Role
	plt, ok := tgs.Platforms[t.Platform]
	if !ok {
		// wait for user reaction to verify
		war, _ := ses.ChannelMessageSend(msg.ChannelID,
			fmt.Sprintf("Warning: Creating a new platform. Make sure you check if a similar one exists.\n"+
				"React to confirm this action. You have %d seconds to confirm.", addTimeout))

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

			select {
			case <-reacted:
			case <-time.After(addTimeout * time.Second):
				reaction <- false
			}

			kill()
		}()

		// check what we got
		if !<-reaction {
			out.Message("Aborting platform creation.")
			return out, nil
		}

		// acknowledge reaction
		ses.ChannelMessageSend(msg.ChannelID, "Creating new platform: "+utils.Code(t.Platform))

		/* role
		// create new role
		drl, err = ses.GuildRoleCreate(msg.GuildID)
		if err != nil {
			return nil, err
		}

		// edit the role
		ses.GuildRoleEdit(msg.GuildID, drl.ID, t.Platform, teal, false, drl.Permissions, true)

		// signal role creation
		ses.ChannelMessageSend(msg.ChannelID, "Creating a new role: "+drl.Mention())
		*/

		// create new platform
		plt = &platform{
			Name:  t.Platform,
			Role:  nil,
			Users: make(map[string]*tag),
		}
		tgs.Platforms[t.Platform] = plt
	}

	/* opt-in pings
	// signal role giving
	ses.ChannelMessageSend(msg.ChannelID, "Giving you the role...")

	// set role, silently fails
	ses.GuildMemberRoleAdd(msg.GuildID, msg.Author.ID, plt.Role.ID)
	*/

	// add tag to platform
	plt.Users[msg.Author.ID] = &tag{
		UID:      msg.Author.ID,
		Username: msg.Author.Username,
		Tag:      argTag,
		Platform: t.Platform,
		PingMe:   false, // opt-in pings
	}

	// set tags
	_, _, err = commands.DBSet(&tgs, tagsKey)
	if err != nil {
		return nil, err
	}

	out.Message("Success! Added tag " + utils.Code(argTag) + " for " + utils.Code(t.Platform))
	return out, nil
}

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

func (t *tagsClean) Roles() []string { return []string{"mod"} }

func (t *tagsClean) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer

	// check if we're cleaning
	if !cleanSemaphore.TryAcquire(1) {
		return nil, ErrCleanSpam
	}
	defer cleanSemaphore.Release(1)

	// lock the db
	commands.DBLock()
	defer commands.DBUnlock()

	// get all tags
	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
		return nil, ErrNoTags
	} else if err != nil {
		return nil, err
	}

	ses.ChannelMessageSend(msg.ChannelID, "Starting cleaning session! This may take a while...")

	// get roles
	/* roles
	roles, err := ses.GuildRoles(msg.GuildID)
	if err != nil {
		return nil, err
	}
	*/

	// cache already seen uids
	checkMap := make(map[string]bool)

	// iterate platforms
	for pname, plt := range tgs.Platforms {
		// clean empty platforms
		if len(plt.Users) == 0 || len(plt.Name) == 0 {
			// remove the role from guild, fails silently
			ses.GuildRoleDelete(msg.GuildID, plt.Role.ID)

			// remove the platform
			delete(tgs.Platforms, pname)
			ses.ChannelMessageSend(msg.ChannelID, "Removed empty platform: "+utils.Code(pname))
			continue
		}

		/* roles
		// check role associated with platform
		exists := false
		for _, rol := range roles {
			if rol.ID == plt.Role.ID {
				exists = true
				break
			}
		}
		if !exists {
			// ASSUME only error is role not existing in server
			// create new role
			var drl *discordgo.Role
			drl, err = ses.GuildRoleCreate(msg.GuildID)
			if err != nil {
				return nil, err
			}

			// edit the role
			ses.GuildRoleEdit(msg.GuildID, drl.ID, pname, teal, false, drl.Permissions, true)
			if err != nil {
				return nil, err
			}

			// add role to platform
			plt.Role = drl
			ses.ChannelMessageSend(msg.ChannelID, "Re-created missing role for platform: "+utils.Code(pname))
		}
		*/

		// check valid users
		for uid, _ := range plt.Users {
			res, ok := checkMap[uid]
			if ok {
				if !res {
					// has been checked and is invalid, remove
					delete(plt.Users, uid)
					ses.ChannelMessageSend(msg.ChannelID, "Removed invalid user: "+uid)
				}
				continue
			}

			// check user
			mem, err := ses.State.Member(msg.GuildID, uid)
			if err != nil {
				mem, err = ses.GuildMember(msg.GuildID, uid)
				if err != nil {
					// couldn't find user, remove tag from db
					delete(plt.Users, uid)
					ses.ChannelMessageSend(msg.ChannelID, "Removed invalid user: "+uid)

					// update cache
					checkMap[uid] = false
					continue
				}
			}

			/* roles
			// update roles
			if usr.PingMe {
				ses.GuildMemberRoleAdd(msg.GuildID, uid, plt.Role.ID)
			} else {
				ses.GuildMemberRoleRemove(msg.GuildID, uid, plt.Role.ID)
			}
			*/

			// update usernames
			plt.Users[uid].Username = mem.User.Username
			ses.ChannelMessageSend(msg.ChannelID, "Updated username for "+uid+" to "+mem.User.Username)

			// update cache
			checkMap[uid] = true
		}
	}

	_, _, err = commands.DBSet(&tgs, tagsKey)
	if err != nil {
		return nil, err
	}

	return commands.NewSimpleSend(msg.ChannelID, "Thanks for waiting, we're all clean now! "+emojiClean), nil
}

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
	for i := range list {
		if i == 6 || i == platLimit+9 {
			list += "+"
		} else {
			list += "-"
		}
	}
	list += "\n"

	// update usernames
	utags := []*tag{}
	for _, utg := range plt.Users {
		mem, err := ses.State.Member(msg.GuildID, utg.UID)
		if err != nil {
			/* expensive
			mem, err = ses.GuildMember(msg.GuildID, utg.UID)
			if err != nil {
				utags = append(utags, nil)
				continue
			}
			*/
			utags = append(utags, nil)
			continue
		}
		utg.Username = mem.User.Username
		utags = append(utags, utg)
	}

	sort.Slice(utags, func(i, j int) bool {
		// move nils to the end
		if utags[i] == nil {
			return false
		}

		if strings.Compare(utags[i].Username, utags[i].Username) < 0 {
			return true
		}
		return false
	})

	// generate output
	for _, utg := range utags {
		if utg == nil {
			// signal invalid users in the db
			list += fmt.Sprintf(fmt.Sprintf("%%-%dt | %%-%ds | %%s\n", 5, userLimit),
				false, "[INVALID]", "!tags clean")
		} else {
			ind := len(utg.Username)
			if len(utg.Username) > userLimit {
				ind = userLimit
			}
			list += fmt.Sprintf(fmt.Sprintf("%%-%dt | %%-%ds | %%s\n", 5, userLimit),
				utg.PingMe, utg.Username[0:ind], utg.Tag)
		}
	}

	return commands.NewSimpleSend(msg.ChannelID, t.Platform+"'s tags:\n"+utils.Block(list)), nil
}

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
	plats := []*platform{}
	for _, plt := range tgs.Platforms {
		plats = append(plats, plt)
	}

	// sort platforms
	sort.Slice(plats, func(i, j int) bool {
		if strings.Compare(plats[i].Name, plats[j].Name) < 0 {
			return true
		}
		return false
	})

	// create message
	list := ""
	for _, plt := range plats {
		list += fmt.Sprintf(fmt.Sprintf("%%-%ds", platLimit), plt.Name)
		list += "|  " + strconv.Itoa(len(plt.Users)) + " tag(s)\n"
	}

	out := "Platforms:\n" + utils.Block(list)
	return commands.NewSimpleSend(msg.ChannelID, out), nil
}

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
	out := commands.NewSend(msg.ChannelID)

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

	pings := ""
	for _, utg := range plt.Users {
		/* slow
		var dusr *discordgo.User
		dusr, err = ses.User(utg.UID)
		if err != nil || !utg.PingMe {
			continue
		}
		pings += " " + dusr.Mention()
		*/
		if utg.PingMe {
			pings += " " + utils.Mention(utg.UID)
		}
	}

	if len(pings) == 0 {
		return out.Message("No one wants " + utils.Code(plt.Name) + " pings."), nil
	}

	pings = utils.Bold(plt.Name) + pings + "\n" + strings.Join(t.Message, " ")

	return out.Message(pings), nil
}

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

	// lock the db
	commands.DBLock()
	defer commands.DBUnlock()

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

	/* roles
	if t.PingMe {
		// TODO: fix this shit
		// set role, silently fails
		ses.GuildMemberRoleAdd(msg.GuildID, msg.Author.ID, plt.Role.ID)
	} else {
		// remove role, silently fails
		ses.GuildMemberRoleRemove(msg.GuildID, msg.Author.ID, plt.Role.ID)
	}
	*/
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

	// lock the db
	commands.DBLock()
	defer commands.DBUnlock()

	// get all tags
	err = commands.DBGet(&tgs, tagsKey, &tgs)
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

	if utg.PingMe {
		// remove role, silently fails
		ses.GuildMemberRoleRemove(msg.GuildID, msg.Author.ID, plt.Role.ID)
	}

	// remove the tag
	delete(plt.Users, msg.Author.ID)
	out.Message("Removed your tag from " + utils.Code(t.Platform))

	if len(plt.Users) == 0 {
		// remove the role from guild, silently fails
		ses.GuildRoleDelete(msg.GuildID, plt.Role.ID)

		// remove the platform
		delete(tgs.Platforms, t.Platform)
		out.Message("Removing empty platform: " + utils.Code(t.Platform))
	}

	_, _, err = commands.DBSet(&tgs, tagsKey)
	if err != nil {
		return nil, err
	}

	return out, nil
}

type tagsUser struct {
	nilCommand
	User []string `arg:"username"`
}

func newTagsUser() *tagsUser { return &tagsUser{} }

func (t *tagsUser) Aliases() []string { return []string{"tags user", "tags view"} }

func (t *tagsUser) Desc() string {
	return "Lists all tags of a user. Use a @ping or a case-insensitive username (not nickname) search." +
		" Empty username will get your own tags."
}

func (t *tagsUser) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer
	var usr *discordgo.User

	if len(t.User) == 0 {
		// get self
		usr = msg.Author
	} else {
		if len(msg.Mentions) > 0 {
			// use mention to get user
			usr = msg.Mentions[0]
		} else {
			// try as uid
			mem, err := ses.GuildMember(msg.GuildID, strings.TrimSpace(t.User[0]))
			if err != nil {
				// use string to search for user
				members, err := ses.GuildMembers(msg.GuildID, "0", guildMemberLimit)
				if err != nil {
					return nil, err
				}

				for _, mem := range members {
					// case-insensitive
					if strings.ToLower(mem.User.Username) == strings.ToLower(t.User[0]) {
						usr = mem.User
					}
				}

				if usr == nil {
					return nil, ErrUserNotFound
				}
			} else {
				usr = mem.User
			}
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

	// sort tags
	sort.Slice(utgs, func(i, j int) bool {
		if strings.Compare(utgs[i].Username, utgs[j].Username) < 0 {
			return true
		}
		return false
	})

	list := fmt.Sprintf(fmt.Sprintf("Ping? | %%-%ds | %%s\n", platLimit), "Platform", "Tag")
	for i := range list {
		if i == 6 || i == platLimit+9 {
			list += "+"
		} else {
			list += "-"
		}
	}
	list += "\n"

	for _, utg := range utgs {
		list += fmt.Sprintf(fmt.Sprintf("%%-%dt | %%-%ds | %%s\n", 5, platLimit),
			utg.PingMe,
			utg.Platform,
			utg.Tag)
	}

	return commands.NewSimpleSend(msg.ChannelID, usr.Username+"'s tags:\n"+utils.Block(list)), nil
}

type tagsModRemove struct {
	nilCommand
	Platform string `arg:"platform"`
}

func newTagsModRemove() *tagsModRemove { return &tagsModRemove{} }

func (t *tagsModRemove) Aliases() []string { return []string{"tags modremove", "tags mod remove"} }

func (t *tagsModRemove) Desc() string { return "Moderator tool to forcibly remove platforms" }

func (t *tagsModRemove) Roles() []string { return []string{"mod"} }

func (t *tagsModRemove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error
	var tgs tagStorer

	// lock the db
	commands.DBLock()
	defer commands.DBUnlock()

	// get all tags
	err = commands.DBGet(&tgs, tagsKey, &tgs)
	if err == commands.ErrDBNotFound {
		return nil, ErrNoTags
	} else if err != nil {
		return nil, err
	}

	// iterate platforms
	for pname, plt := range tgs.Platforms {
		if pname == t.Platform {
			// remove the role from guild, fails silently
			ses.GuildRoleDelete(msg.GuildID, plt.Role.ID)

			// remove the platform
			delete(tgs.Platforms, pname)

			// commit changes
			_, _, err = commands.DBSet(&tgs, tagsKey)
			if err != nil {
				return nil, err
			}
			return commands.NewSimpleSend(msg.ChannelID, "Removed platform: "+utils.Code(pname)), nil
		}
	}

	return nil, ErrNoPlatform
}
