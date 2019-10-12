package utils

import (
	"reflect"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Bold encloses string in bold tags
func Bold(s string) string {
	return "**" + s + "**"
}

// Block encloses string in code block tags
func Block(s string) string {
	return "```\n" + s + "\n```"
}

// Code encloses string in code tags
func Code(s string) string {
	return "`" + s + "`"
}

// Italics encloses string in italics tags
func Italics(s string) string {
	return "*" + s + "*"
}

// Spoil encloses string in spoiler tags
func Spoil(s string) string {
	return "||" + s + "||"
}

// Under encloses string in underline tags
func Under(s string) string {
	return "__" + s + "__"
}

// Mention encloses the string in mention tags
func Mention(s string) string {
	return "<@!" + s + ">"
}

// Reverse reverses a string, assuming ascii encoding
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// StrLen Recursively searches for strings and counts up the total length
func Strlen(e interface{}) int {
	count := 0
	// type assert e
	ev, ok := e.(reflect.Value)
	if !ok {
		ev = reflect.ValueOf(e)
	}

	switch ev.Kind() {
	case reflect.String:
		// count string lengths
		count += len(ev.String())

	case reflect.Ptr:
		// unroll pointers
		elv := ev.Elem()
		if !elv.IsValid() {
			return 0
		}
		count += Strlen(elv)

	case reflect.Struct:
		// iterate over fields
		for i := 0; i < ev.NumField(); i++ {
			switch f := ev.Field(i); f.Kind() {
			case reflect.String:
				// count string field lengths
				count += len(f.String())

			case reflect.Ptr:
				// unroll pointers
				el := f.Elem()
				if !el.IsValid() {
					return 0
				}
				count += Strlen(el)

			case reflect.Struct:
				// recurse over struct fields
				count += Strlen(f.Interface())

			default:
				// do nothing
			}
		}
	default:
		// do nothing
	}
	return count
}

// MsgHasRoles Checks if the author has the required roles
func MsgHasRoles(ses *discordgo.Session, msg *discordgo.Message, roles []string) (bool, error) {
	if len(roles) == 0 {
		return true, nil
	}

	// Get member
	member, err := ses.State.Member(msg.GuildID, msg.Author.ID)
	if err != nil {
		member, err = ses.GuildMember(msg.GuildID, msg.Author.ID)
		if err != nil {
			return false, err
		}
	}

	// Get guild roles
	groles, err := ses.GuildRoles(msg.GuildID)
	if err != nil {
		return false, err
	}

	// Get roles required
	rolesrequired := []string{}
	for _, r := range roles {
		for _, gr := range groles {
			if strings.ToLower(gr.Name) == r {
				rolesrequired = append(rolesrequired, gr.ID)
			}
		}
	}

	// Check member roles
	mroles := member.Roles
	for _, rr := range rolesrequired {
		for _, mr := range mroles {
			if mr == rr {
				return true, nil
			}
		}
	}

	return false, nil
}

// MsgInChannels Checks if message was sent in the required channels
func MsgInChannels(s *discordgo.Session, m *discordgo.Message, channels []string) (bool, error) {
	if len(channels) == 0 {
		return true, nil
	}

	cha, err := s.Channel(m.ChannelID)
	if err != nil {
		return false, err
	}

	for _, c := range channels {
		if c == cha.Name {
			return true, nil
		}
	}

	return false, nil
}
