package commands_test

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"

	. "github.com/unswpcsoc/PCSocBot/commands"
)

/* preamble */

type BadPing struct {
	Name   string   `arg:"name"`
	Age    int      `arg:"age"`
	Rest   []string `arg:"rest"`
	IsCool bool     `arg:"cool?"`
}

func NewBadPing() *BadPing { return &BadPing{} }

func (p *BadPing) Aliases() []string { return []string{"ping", "ping pong"} }

func (p *BadPing) Desc() string { return "BadPing!" }

func (p *BadPing) Roles() []string { return nil }

func (p *BadPing) Chans() []string { return nil }

func (p *BadPing) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	return nil, nil
}

type Ping struct {
	Name   string   `arg:"name"`
	Age    int      `arg:"age"`
	IsCool bool     `arg:"cool?"`
	Rest   []string `arg:"rest"`
}

func NewPing() *Ping { return &Ping{} }

func (p *Ping) Aliases() []string { return []string{"ping", "ping pong"} }

func (p *Ping) Desc() string { return "Ping!" }

func (p *Ping) Roles() []string { return nil }

func (p *Ping) Chans() []string { return nil }

func (p *Ping) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*CommandSend, error) {
	return nil, nil
}

/* tests */

// TestArgFill gives ArgFill some args
// and verifies that ArgFill will:
// - Fill args correctly when possible
// - Throws error when string conversion fails
// - Throws error when not enough args are supplied
// - TODO: Panics when var args are misused by the programmer
func TestArgFill(t *testing.T) {
	var err error
	var args []string
	exp := NewPing()
	got := NewPing()

	// fill with nothing
	args = []string{}
	err = FillArgs(got, args)
	if err != ErrNotEnoughArgs {
		t.Errorf("ArgFill(%v, %v) threw error: %v\nexpected error: %v", got, args, err, ErrNotEnoughArgs)
	}

	// fill with incomplete args
	args = []string{"bob", "42"}
	err = FillArgs(got, args)
	if err != ErrNotEnoughArgs {
		t.Errorf("ArgFill(%v, %v) threw error: %v\nexpected error: %v", got, args, err, ErrNotEnoughArgs)
	}

	// fill with all args
	exp.Name = "bob"
	exp.Age = 42
	exp.IsCool = true
	exp.Rest = []string{"bob", "is", "cool"}

	args = []string{"bob", "42", "true", "bob", "is", "cool"}
	err = FillArgs(got, args)

	if err != nil {
		t.Errorf("ArgFill(%#v, %v)\nthrew error: %v", NewPing(), args, err)
	}
	if !reflect.DeepEqual(got, exp) {
		t.Errorf("ArgFill(%#v, %v)\nset %#v\nwant %#v", NewPing(), args, got, exp)
	}

	// test panic
	pan := NewBadPing()
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Caught panic on bad var args: %v", r)
			return
		}
	}()

	// fill badly-structured command with args
	args = []string{"bob", "42", "bob", "is", "cool", "true"}
	err = FillArgs(pan, args)
	t.Errorf("ArgFill(%#v, %v)\nDidn't panic with bad var args placement!", NewBadPing(), args)
}
