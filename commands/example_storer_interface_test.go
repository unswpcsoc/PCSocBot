package commands_test

import (
	"fmt"

	"github.com/unswpcsoc/PCSocBot/commands"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`

	// DO NOT DO THIS
	//c string `json:"unexported"`
}

// Index implements commands.Storer
// for thing
func (p *Person) Index() string {
	return "person"
}

func Example() {
	commands.DBOpen(":memory:")
	defer commands.DBClose()

	exp := Person{
		Name: "Bob",
		Age:  42,
	}

	fmt.Printf("exp = %#v\n", exp)

	// Our commands.Storer can be entered into the db by calling the interface "method" DBSet
	commands.DBSet(&exp, "0")

	// Our commands.Storer can be accessed from the db by calling the interface "method" DBGet
	var got Person
	commands.DBGet(&Person{}, "0", &got)
	fmt.Printf("got = %#v\n", got)

	// Output:
	// exp = commands_test.Person{Name:"Bob", Age:42}
	// got = commands_test.Person{Name:"Bob", Age:42}
}
