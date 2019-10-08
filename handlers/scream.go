package handlers

import (
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/utils"
)

type scream struct {
	nilCommand
}

func newScream() *scream { return &scream{} }

func (s *scream) Aliases() []string { return []string{"scream"} }

func (s *scream) Desc() string { return "AAAAAAAAAAAAAAAA" }

func (s *scream) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// seed randomness
	rand.Seed(time.Now().UnixNano())

	// roll 50-sided die
	var out string
	switch rand.Intn(50) {
	case 0: // HHHhhhh
		out = utils.Reverse(trailCaps("h"))

	case 1: // waaAAAA
		out = "w" + trailCaps("a")

	case 2: // AAAaaaa
		out = trailCaps("a")

	case 3: // eeeEEEE
		out = trailCaps("e")

	default: // regular homerow mashing, and then some
		var screams = "asdfghjklbn"
		var starts = "fjk"

		if rand.Intn(2) == 0 {
			// change to uppercase
			screams = strings.ToUpper(screams)
			starts = strings.ToUpper(starts)
		}

		// generate start
		out = string(starts[rand.Intn(len(starts))])

		// generate the rest
		length := 5 + rand.Intn(10)
		for i := 0; i < length; i++ {
			// pick new letter
			app := screams[rand.Intn(len(screams))]

			// make sure no two letters are repeated
			// unless it's a h
			// there is a low probability that this will cause an 'infinite' loop
			if isSameLetter(app, out[len(out)-1]) && !isSameLetter(app, byte('h')) {
				length++
				continue
			} else {
				out += string(app)
			}
		}
	}
	// yeet the scream
	return commands.NewSimpleSend(msg.ChannelID, out), nil
}

// expects a single character, repeats it with varying trailing caps
func trailCaps(str string) string {
	rand.Seed(time.Now().UnixNano())
	return strings.Repeat(str, rand.Intn(10)+5) + strings.ToUpper(strings.Repeat(str, rand.Intn(10)))
}

// checks if a byte is alphabet equivalent to another, ignoring case
func isSameLetter(lhs, rhs byte) bool {
	// pls don't do this ever
	return lhs == strings.ToLower(string(rhs))[0] || lhs == strings.ToUpper(string(rhs))[0]
}
