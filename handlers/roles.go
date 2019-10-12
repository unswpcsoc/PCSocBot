package handlers

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/internal/utils"
)

type role struct {
	nilCommand
	names []string
	desc  string
	rol   string
}

func newRole(rol string) *role {
	return &role{
		names: []string{strings.ToLower(rol)},
		desc:  "Gives user the " + rol + " role.",
		rol:   strings.ToLower(rol),
	}
}

func (r *role) Aliases() []string { return r.names }

func (r *role) Desc() string { return r.desc }

func (r *role) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	var err error

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
	guildroles, err := ses.GuildRoles(msg.GuildID)
	if err != nil {
		return nil, err
	}

	for _, role := range guildroles {
		if strings.ToLower(r.rol) == strings.ToLower(role.Name) {
			roleID = role.ID
			break
		}
	}

	if len(roleID) == 0 {
		return nil, errors.New("no such role " + utils.Code(r.rol) + " in guild: " + msg.GuildID + "\n")
	}

	for _, role := range mem.Roles {
		if roleID == role {
			has = true
		}
	}

	// Add / remove the role accordingly
	if has {
		err = ses.GuildMemberRoleRemove(msg.GuildID, msg.Author.ID, roleID)
		if err != nil {
			return nil, err
		}

		return commands.NewSimpleSend(msg.ChannelID, mem.Mention()+" is no longer a "+r.rol), nil
	}

	err = ses.GuildMemberRoleAdd(msg.GuildID, msg.Author.ID, roleID)
	if err != nil {
		return nil, err
	}

	return commands.NewSimpleSend(msg.ChannelID, mem.Mention()+" is now a "+r.rol), nil
}
