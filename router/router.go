package router

import (
	"sort"
	"strings"

	comm "github.com/unswpcsoc/PCSocBot/commands"
)

// Leaf is a leaf of the command router tree.
type Leaf struct {
	Command comm.Command
	Leaves  map[string]*Leaf
}

// NewLeaf returns a Leaf with the command.
func NewLeaf(com comm.Command) *Leaf {
	return &Leaf{
		Command: com,
		Leaves:  make(map[string]*Leaf),
	}
}

// Router routes a command string to a command.
type Router struct {
	Routes *Leaf
}

// NewRouter returns a new Router structure.
func NewRouter() *Router {
	return &Router{NewLeaf(nil)}
}

// AddCommand adds command-string mapping
func (r *Router) AddCommand(com comm.Command) {
	if com == nil || len(com.Aliases()) == 0 || r.Routes == nil {
		return
	}

	for _, str := range com.Aliases() {
		argv := strings.Split(str, " ")

		// Search all known leaves
		curr := r.Routes
		for {
			next, found := curr.Leaves[argv[0]]
			if !found {
				// New branching
				break
			}
			curr = next
			argv = argv[1:]
			if len(argv) == 0 {
				// All argv match
				break
			}
		}

		// Add new leaves for remaining args
		for len(argv) > 0 {
			curr.Leaves[argv[0]] = NewLeaf(nil)
			curr = curr.Leaves[argv[0]]
			argv = argv[1:]
		}

		// Assign command to the final leaf
		curr.Command = com
	}
}

// Route routes to handler from string.
// Returns the command and the number of matched args.
// e.g.
//	   // r has a route through "example"->"command"->"string"
//     com, ind := r.Route([]string{"example", "command", "string", "with", "args"})
//
// `com` will contain the command at "string" leaf
// `ind` will be 3
func (r *Router) Route(argv []string) (comm.Command, int) {
	if r.Routes == nil || len(argv) == 0 {
		return nil, 0
	}

	// iterate through routes
	i := 0
	curr := r.Routes
	var prev *Leaf = nil
	var ok bool
	for i = 0; i < len(argv); i++ {
		curr, ok = curr.Leaves[argv[i]]
		if !ok {
			break
		}
		prev = curr
	}

	if prev == nil {
		return nil, i
	}

	return prev.Command, i
}

// ToSlice searches the tree and populates a slice of Commands
// sorted by the first alias name
//
// Duplicates are removed in case you were wondering
func (r *Router) ToSlice() []comm.Command {
	var commands = make(map[comm.Command]bool)
	var doToSlice func(*Leaf)

	doToSlice = func(curr *Leaf) {
		if curr == nil {
			return
		}
		if curr.Command != nil {
			commands[curr.Command] = true
		}
		for _, l := range curr.Leaves {
			doToSlice(l)
		}
		return
	}
	doToSlice(r.Routes)

	keys := []comm.Command{}
	for key := range commands {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Aliases()[0] < keys[j].Aliases()[0]
	})
	return keys
}

// ToStringSlice wraps ToSlice and populates a strings slice of aliases
func (r *Router) ToStringSlice() []string {
	out := []string{}
	for _, cmd := range r.ToSlice() {
		for _, ali := range cmd.Aliases() {
			out = append(out, ali)
		}
	}
	return out
}
