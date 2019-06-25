package router

import (
	"strings"

	comm "github.com/unswpcsoc/PCSocBot/commands"
)

// Leaf is a leaf of the command router tree.
type Leaf struct {
	command comm.Command
	leaves  map[string]*Leaf
}

// NewLeaf returns a Leaf with the command.
func NewLeaf(com comm.Command) *Leaf {
	return &Leaf{
		command: com,
		leaves:  make(map[string]*Leaf),
	}
}

// Router routes a command string to a command.
type Router struct {
	routes *Leaf
}

// NewRouter returns a new Router structure.
func NewRouter() Router {
	return Router{NewLeaf(nil)}
}

// Addcommand adds command-string mapping
func (r *Router) Addcommand(com comm.Command) {
	if com == nil || len(com.Aliases()) == 0 || r.routes == nil {
		return
	}

	for _, str := range com.Aliases() {
		argv := strings.Split(str, " ")

		// Search all known leaves
		curr := r.routes
		for {
			next, found := curr.leaves[argv[0]]
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
			curr.leaves[argv[0]] = NewLeaf(nil)
			curr = curr.leaves[argv[0]]
			argv = argv[1:]
		}

		// Assign command to the final leaf
		curr.command = com
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
	if r.routes == nil || len(argv) == 0 {
		return nil, 0
	}

	// iterate through routes
	i := 0
	curr := r.routes
	var prev *Leaf = nil
	var ok bool
	for i = 0; i < len(argv); i++ {
		curr, ok = curr.leaves[argv[i]]
		if !ok {
			break
		}
		prev = curr
	}

	if prev == nil {
		return nil, i
	}

	return prev.command, i
}
