// Package commands implements a command interface for PCSocBot
// with helper structs and funcs for sending discordgo messages,
// and high-level abstractions of buntdb
package commands

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/internal/utils"
)

const (
	// Prefix is the prefix for commands
	Prefix = "!"
	// MessageLimit is the character limit for messages
	MessageLimit = 2000
)

var (
	// ErrSendLimit means the message was too long
	ErrSendLimit = errors.New("message exceeds send limit of 2000 characters")
	// ErrNotEnoughArgs means the user did not provide enough arguments to the command
	ErrNotEnoughArgs = errors.New("not enough arguments provided")
)

// Command is the interface that all commands implement.
type Command interface {
	Aliases() []string      // Aliases
	Desc() string           // Description
	Subcommands() []Command // Slice of subcommands, if any
	Roles() []string        // Roles required
	Chans() []string        // Channels required

	MsgHandle(*discordgo.Session, *discordgo.Message) (*CommandSend, error) // Handler for MessageCreate event
}

// CommandSend is a helper struct that buffers things commands need to send.
type CommandSend struct {
	data      []*discordgo.MessageSend
	channelid string
}

// NewSend Returns a send struct.
func NewSend(cid string) *CommandSend {
	return &CommandSend{
		[]*discordgo.MessageSend{},
		cid,
	}
}

// NewSimpleSend Returns a send struct with the message content filled in.
func NewSimpleSend(cid string, msg string) *CommandSend {
	send := &discordgo.MessageSend{
		Content: msg,
	}
	return &CommandSend{
		data:      []*discordgo.MessageSend{send},
		channelid: cid,
	}
}

// Message Adds another simple message to be sent.
func (c *CommandSend) Message(msg string) *CommandSend {
	send := &discordgo.MessageSend{
		Content: msg,
	}
	c.data = append(c.data, send)
	return c
}

// Embed Adds an embed-only message to be sent.
func (c *CommandSend) Embed(emb *discordgo.MessageEmbed) *CommandSend {
	send := &discordgo.MessageSend{
		Embed: emb,
	}
	c.data = append(c.data, send)
	return c
}

// MessageSend Adds a discordgo MessageSend.
func (c *CommandSend) MessageSend(send *discordgo.MessageSend) *CommandSend {
	c.data = append(c.data, send)
	return c
}

// Send Sends the messages a command returns while also checking message length
func (c *CommandSend) Send(s *discordgo.Session) error {
	// Get the stuff out of BeegYoshi and send it into the server
	for _, data := range c.data {
		if utils.Strlen(data) > MessageLimit {
			return fmt.Errorf("Send: following message exceeds limit\n%#v", data)
		}
		s.ChannelMessageSendComplex(c.channelid, data)
	}
	return nil
}

/* usage generation */

// GetUsage generates the usage message from a Command in the following format
//  !alias0 (type0) __arg0__ (type1) __arg1__ ...
//  description of command
//  __Aliases__ | !alias1 | !alias2 ...
func GetUsage(c Command) (usage string) {
	v := reflect.ValueOf(c)

	if v.Kind() == reflect.Ptr {
		// unroll pointer
		v = v.Elem()
		if !v.IsValid() {
			panic(fmt.Sprintf("GetUsage: %v is not a valid pointer\n", v))
		}
	}

	if v.Kind() != reflect.Struct {
		panic(fmt.Sprintf("GetUsage: %v is not a struct\n", v))
	}

	// command alias
	names := c.Aliases()
	usage = utils.Bold("!" + names[0])

	// parse struct fields with arg tags
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)

		tag, ok := f.Tag.Lookup("arg")
		if !ok {
			continue
		}

		// kind switch to parse into human-readable
		// names for types
		var tName string
		switch f.Type.Kind() {
		case reflect.Bool:
			tName = "true/false"
		case reflect.Int:
			tName = "number"
		case reflect.String:
			tName = "word"
		case reflect.Array, reflect.Slice:
			tName = "multiple "
			switch f.Type.Elem().Kind() {
			case reflect.Bool:
				tName += "true/false"
			case reflect.Int:
				tName += "numbers"
			case reflect.String:
				tName += "words"
			}
		default:
			tName = f.Type.Name()
		}

		usage += " (" + tName + ") " + utils.Under(tag)
	}

	// description
	usage += "\n" + c.Desc()

	// aliases
	if len(names) > 1 {
		usage += "\n" + utils.Under("Aliases")
		for _, name := range names[1:] {
			usage += " | !" + name
		}
	}

	// subcommands
	if len(c.Subcommands()) > 0 {
		usage += "\n" + utils.Under("Subcommands")
		for _, sc := range c.Subcommands() {
			usage += " | !" + sc.Aliases()[0]
		}
	}

	if len(usage) > MessageLimit {
		panic("command is too damn big!")
	}

	return usage
}

// FillArgs tries to fill the given command's struct fields with the args given
//
// FillArgs will return a strconv error if types cannot be matched
// and will panic if there are unexported arg fields or if variable args are done incorrectly
// or if input is generally messed up
func FillArgs(c Command, args []string) error {
	var err error
	var val reflect.Value
	val = reflect.ValueOf(c)

	if val.Kind() == reflect.Ptr {
		// unroll pointer
		val = val.Elem()
		if !val.IsValid() {
			panic(fmt.Sprintf("FillArgs: %#v is not valid\n", val))
		}
	}

	if val.Kind() != reflect.Struct {
		panic(fmt.Sprintf("FillArgs: %#v is not a struct\n", val))
	}

	// get arg fields as slice of values
	argFields := []reflect.Value{}
	for i := 0; i < val.NumField(); i++ {
		ft := val.Type().Field(i)
		fv := val.Field(i)
		_, ok := ft.Tag.Lookup("arg")
		if !ok {
			continue
		}
		if !fv.CanSet() {
			panic("FillArgs: using unexported field with arg tag")
		}
		argFields = append(argFields, fv)
	}

	if len(argFields) == 0 {
		return nil
	}

	if len(args) < len(argFields) {
		// check var args in the last slot
		if !(len(args) == len(argFields)-1 && argFields[len(argFields)-1].Kind() == reflect.Slice) {
			return ErrNotEnoughArgs
		}
	}

	// iterate through arg fields
	argIndex := 0
	for i, fv := range argFields {
		// make sure we don't hit invalid indexes
		if i == len(args) {
			break
		}

		// kind switch for field types
		switch fv.Kind() {
		case reflect.String:
			fv.SetString(args[argIndex])

		case reflect.Int:
			var got int
			got, err = strconv.Atoi(args[argIndex])
			if err != nil {
				return err
			}
			fv.SetInt(int64(got))

		case reflect.Bool:
			var got bool
			got, err = strconv.ParseBool(args[argIndex])
			if err != nil {
				return err
			}
			fv.SetBool(got)

		case reflect.Array, reflect.Slice:
			// check slice is last arg field
			if i+1 != len(argFields) {
				panic("FillArgs: variable-length arg but is not the final arg field")
			}

			// make new slice value of slice field's element type
			elemType := fv.Type().Elem()
			sv := reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0)

			// kind switch for slice elem types
			switch elemType.Kind() {
			case reflect.String:
				for ; argIndex < len(args); argIndex++ {
					sv = reflect.Append(sv, reflect.ValueOf(args[argIndex]))
				}
			case reflect.Int:
				for ; argIndex < len(args); argIndex++ {
					var got int
					got, err = strconv.Atoi(args[argIndex])
					if err != nil {
						return err
					}
					sv = reflect.Append(sv, reflect.ValueOf(got))
				}
			case reflect.Bool:
				for ; argIndex < len(args); argIndex++ {
					var got bool
					got, err = strconv.ParseBool(args[argIndex])
					if err != nil {
						return err
					}
					sv = reflect.Append(sv, reflect.ValueOf(got))
				}
			default:
				panic("FillArgs: var-arg field cannot handle elem type " + elemType.Kind().String())
			}

			fv.Set(sv.Slice(0, sv.Len()))
			return nil

		default:
			panic("FillArgs: arg field cannot handle type " + fv.Kind().String())
		}

		// continue along arg
		argIndex++
	}
	return nil
}

// CleanArgs cleans arg-fields from commands after they've been handled
//
// This should be called after your Command is done handling the message.
//
// Also should be called on Command creation if you don't trust your programmers
func CleanArgs(c Command) {
	var val reflect.Value
	val = reflect.ValueOf(c)

	if val.Kind() == reflect.Ptr {
		// unroll pointer
		val = val.Elem()
		if !val.IsValid() {
			panic(fmt.Sprintf("CleanArgs: %#v is not valid\n", val))
		}
	}

	if val.Kind() != reflect.Struct {
		panic(fmt.Sprintf("CleanArgs: %#v is not a struct\n", val))
	}

	// iterate over arg fields and zero them
	for i := 0; i < val.NumField(); i++ {
		ft := val.Type().Field(i)
		fv := val.Field(i)
		_, ok := ft.Tag.Lookup("arg")
		if !ok {
			continue
		}
		if !fv.CanSet() {
			panic("FillArgs: using unexported field with arg tag")
		}

		// TODO: default arg handling goes here
		// zero out field
		fv.Set(reflect.Zero(fv.Type()))
	}
}
