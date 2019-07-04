package router_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"

	comm "github.com/unswpcsoc/PCSocBot/commands"
	. "github.com/unswpcsoc/PCSocBot/router"
)

// signal testing
// https://stackoverflow.com/questions/43409919/init-function-breaking-unit-tests
// not using this anymore, using different package for testing
// interesting read nonetheless

/* preamble */

type Example struct {
	names []string
	desc  string
}

func NewExample() *Example { return &Example{} }

func (e *Example) Aliases() []string { return []string{"example", "an extended command string"} }

func (e *Example) Desc() string { return "Example!" }

func (e *Example) Roles() []string { return nil }

func (e *Example) Chans() []string { return nil }

func (e *Example) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*comm.CommandSend, error) {
	return nil, nil
}

type Example2 struct {
	names []string
	desc  string
}

func NewExample2() *Example2 { return &Example2{} }

func (e *Example2) Aliases() []string {
	return []string{"another example", "an extended command string 2"}
}

func (e *Example2) Desc() string { return "Example2!" }

func (e *Example2) Roles() []string { return nil }

func (e *Example2) Chans() []string { return nil }

func (e *Example2) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*comm.CommandSend, error) {
	return nil, nil
}

/* tests */

func TestMain(m *testing.M) {
	// run tests
	os.Exit(m.Run())
}

func TestAddcommand(t *testing.T) {
	// init router
	router := NewRouter()

	// create command
	exp := NewExample()

	// add to router
	router.Addcommand(exp)

	// assert single route made
	r1 := "example"
	got, ok := router.Routes.Leaves[r1]
	if !ok {
		t.Errorf("%s: no route made for %s\n", t.Name(), r1)
	}
	if got.Command != exp {
		t.Errorf("%s: made route to %v, expected %v\n", t.Name(), got.Command, exp)
	}

	// assert lengthy route made
	r2 := []string{"an", "extended", "command", "string"}
	curr := router.Routes.Leaves
	for _, str := range r2 {
		got, ok = curr[str]
		if !ok {
			t.Errorf("%s: no route made for %v\n", t.Name(), r2)
		}
		curr = got.Leaves
	}
	if got.Command != exp {
		t.Errorf("%s: made route to %v, expected %v\n", t.Name(), got.Command, exp)
	}
}

func TestRoute(t *testing.T) {
	// init router
	router := NewRouter()

	// create command
	exp := NewExample()

	// add simple route manually
	expl := NewLeaf(exp)
	router.Routes.Leaves["example"] = expl

	// assert simple routing works
	got, ind := router.Route([]string{"example"})
	if ind == 0 {
		t.Errorf("%s: route did not find anything\n", t.Name())
	}
	if got != exp {
		t.Errorf("%s: got command %v, expected %v\n", t.Name(), got, exp)
	}

	// add lengthy route manually
	exp2 := NewLeaf(exp)
	router.Routes.Leaves["an"] = NewLeaf(nil)
	router.Routes.Leaves["an"].Leaves["extended"] = NewLeaf(nil)
	router.Routes.Leaves["an"].Leaves["extended"].Leaves["command"] = NewLeaf(nil)
	router.Routes.Leaves["an"].Leaves["extended"].Leaves["command"].Leaves["string"] = exp2

	// assert lengthy routing works
	got, ind = router.Route([]string{"an", "extended", "command", "string", "with", "some", "args"})
	if ind == 0 {
		t.Errorf("%s: route did not find anything\n", t.Name())
	}
	if got != exp {
		t.Errorf("%s: got command %v, expected %v\n", t.Name(), got, exp)
	}
	if ind != 4 {
		t.Errorf("%s: got index %v, expected %v\n", t.Name(), ind, 4)
	}

	// test unregistered commands
	got, ind = router.Route([]string{"help"})
	expind := 0
	if ind != 0 {
		t.Errorf("%s: index %v, expected %v\n", t.Name(), ind, expind)
	}

	if got != nil {
		t.Errorf("%s: got %v, expected %v\n", t.Name(), got, nil)
	}
}

func TestAddRoute(t *testing.T) {
	// init router
	router := NewRouter()

	// create command
	exp := NewExample()

	// add simple route
	router.Addcommand(exp)

	// assert simple routing works
	got, ind := router.Route([]string{"example"})
	if ind == 0 {
		t.Errorf("%s: route did not find anything\n", t.Name())
	}
	if got != exp {
		t.Errorf("%s: got %#v, expected %#v\n", t.Name(), got, exp)
	}

	// assert lengthy routing works
	got, ind = router.Route([]string{"an", "extended", "command", "string"})
	if ind == 0 {
		t.Errorf("%s: route did not find anything\n", t.Name())
	}
	if got != exp {
		t.Errorf("%s: got %#v, expected %#v\n", t.Name(), got, exp)
	}
}

func TestToSlice(t *testing.T) {
	// init router
	router := NewRouter()

	// create commands
	router.Addcommand(NewExample())
	router.Addcommand(NewExample2())

	// get slice
	exp := []comm.Command{&Example{}, &Example2{}}
	got := router.ToSlice()

	if !reflect.DeepEqual(got, exp) {
		t.Errorf("%s: got %#v\nexpected %#v", t.Name(), got, exp)
	}
}
