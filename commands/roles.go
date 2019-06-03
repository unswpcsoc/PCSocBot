package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
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
		role:  role,
	}
}

func (r *Role) Names() []string {
	return r.names
}

func (r *Role) Desc() string {
	return r.desc
}

func (r *Role) Roles() []string {
	return nil
}

func (r *Role) Chans() []string {
	return nil
}

func (r *Role) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	member, _ := ses.GuildMember(msg.GuildID, msg.Author.ID)

	// Find the role ID for this role name
	var roleID string
	guildRoles, err := ses.GuildRoles(msg.GuildID)
	if err != nil {
	    return nil, err
	}
	for _, role := range guildRoles {
		if r.role == strings.ToLower(role.Name) {
			roleID = role.ID
			break
		}
	}

	// Check if the user has the role
	hasRole := false
	for _, role := range member.Roles {
		if roleID == role {
			hasRole = true
		}
	}

	// Add / remove the role accordingly
	if hasRole {
		ses.GuildMemberRoleRemove(msg.GuildID, msg.Author.ID, roleID)
		return NewSimpleSend(msg.ChannelID, member.Mention() + " is no longer a " + r.role), nil
	} else {
		ses.GuildMemberRoleAdd(msg.GuildID, msg.Author.ID, roleID)
		return NewSimpleSend(msg.ChannelID, member.Mention() + " is now a " + r.role), nil
	}
}
