package handlers

import (
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/internal/utils"
)

type scream struct {
	nilCommand
}

func newScream() *scream { return &scream{} }

func (s *scream) Aliases() []string { return []string{"scream"} }

func (s *scream) Desc() string { return "AAAAAAAAAAAAAAAA" }

func (s *scream) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// seed randomness every run
	rand.Seed(time.Now().UnixNano())

	// roll 20-sided die
	var out string
	switch rand.Intn(20) {
	case 0: // HHHhhhh
		out = utils.Reverse(trailCaps("h"))

	case 1: // NmmMMMM
		out = "Nmm" + trailCaps("M")

	case 2: // eeeEEEE
		out = trailCaps("e")

	case 3: // aaaaAAA
		out = trailCaps("a")

	case 4: // AAAaaaa
		out = utils.Reverse(trailCaps("a"))

	case 5: // AAAAAAA
		out = strings.Repeat("A", rand.Intn(500)+1)

	default: // regular homerow mashing, and then some
		var screams = "asdfghjklbn; "
		var starts = "fjk"
		var spaceCount = 0
		var semiCount = 0

		if rand.Intn(2) == 0 {
			// change to uppercase
			screams = strings.ToUpper(screams)
			starts = strings.ToUpper(starts)
		}

		// generate start
		out = string(starts[rand.Intn(len(starts))])

		// generate the rest
		length := 10 + rand.Intn(20)
		for i := 0; i < length; i++ {
			// pick new letter
			app := screams[rand.Intn(len(screams))]

			// make sure no two letters are repeated
			// unless it's a h
			// there is a low probability that this will cause an 'infinite' loop
			if isSameLetter(app, out[len(out)-1]) && !isSameLetter(app, byte('h')) {
				length++
				continue
			} else if isSameLetter(app, byte(' ')) {
				// skip if we have already done a space
				if spaceCount == 1 {
					length++
					continue
				}
				out += string(app)
				spaceCount++
			} else if isSameLetter(app, byte(';')) {
				// skip if we have already done 2 semicolons
				if semiCount == 2 {
					length++
					continue
				}
				out += string(app)
				semiCount++
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
	return strings.Repeat(str, rand.Intn(20)+5) + strings.ToUpper(strings.Repeat(str, rand.Intn(20)))
}

// checks if a byte is alphabet equivalent to another, ignoring case
func isSameLetter(lhs, rhs byte) bool {
	// pls don't do this ever
	return lhs == strings.ToLower(string(rhs))[0] || lhs == strings.ToUpper(string(rhs))[0]
}
