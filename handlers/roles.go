package handlers

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"

	. "github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/utils"
)

type Role struct {
	names []string
	desc  string
	role  string
}

func NewRole(role string) *Role {
	return &Role{
		names: []string{strings.ToLower(role)},
		desc:  "Gives user the " + role + " role.",
		role:  strings.ToLower(role),
	}
}

func (r *Role) Aliases() []string { return r.names }

func (r *Role) Desc() string { return r.desc }

func (r *Role) Subcommands() []Command { return nil }

func (r *Role) Roles() []string { return nil }

func (r *Role) Chans() []string { return nil }

func (r *Role) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	mem, err := ses.State.Member(msg.GuildID, msg.Author.ID)
	if err != nil {
		mem, err = ses.GuildMember(msg.GuildID, msg.Author.ID)
		if err != nil {
			return nil, err
		}
	}

	// Check if user has role
	roleID := ""
	has := false
	guildRoles, err := ses.GuildRoles(msg.GuildID)
	if err != nil {
		return nil, err
	}

	for _, role := range guildRoles {
		if strings.ToLower(r.role) == strings.ToLower(role.Name) {
			roleID = role.ID
			break
		}
	}

	if len(roleID) == 0 {
		return nil, errors.New("no such role " + utils.Code(r.role) + " in guild: " + msg.GuildID + "\n")
	}

	for _, role := range mem.Roles {
		if roleID == role {
			has = true
		}
	}

	// Add / remove the role accordingly
	if has {
		err := ses.GuildMemberRoleRemove(msg.GuildID, msg.Author.ID, roleID)
		if err != nil {
			return nil, err
		}

		return NewSimpleSend(msg.ChannelID, mem.Mention()+" is no longer a "+r.role), nil
	} else {
		err := ses.GuildMemberRoleAdd(msg.GuildID, msg.Author.ID, roleID)
		if err != nil {
			return nil, err
		}

		return NewSimpleSend(msg.ChannelID, mem.Mention()+" is now a "+r.role), nil
	}
}
