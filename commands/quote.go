package commands

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/utils"
)

/* In this file:

quote
quote list
quote pending
quote add
quote approve
quote remove
quote reject

*/

const (
	/* keys for quotes */
	KeyPending = "pending"
	KeyQuotes  = "approve"

	/* limits */
	QuoteLineLimit = 50
)

var (
	ErrQuoteIndex = errors.New("index not valid")
	ErrQuoteEmpty = errors.New("list is empty")
	ErrQuoteArgs  = errors.New("not enough args")
)

/* Storer: quotes */

// quotes implements the Storer interface
type quotes struct {
	List []string
	Last int
}

func (q *quotes) Index() string {
	return "quotes"
}

/* quote */

type Quote struct {
	Index int `arg:"index"`
}

func NewQuote() *Quote { return &Quote{} }

func (q *Quote) Aliases() []string { return []string{"quote"} }

func (q *Quote) Desc() string { return "Get a quote at given index. No args gives a random quote." }

func (q *Quote) Roles() []string { return nil }

func (q *Quote) Chans() []string { return nil }

func (q *Quote) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Get quotes
	var quo quotes
	err := DBGet(&quotes{}, KeyQuotes, &quo)
	if err == ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// Check args
	var ind int
	if len(args) == 0 {
		// Gen random number
		rand.Seed(time.Now().UnixNano())
		ind = rand.Intn(len(quo.List))
	} else {
		// Try get index
		ind, err = strconv.Atoi(args[0])
		if err != nil || ind > quo.Last || ind < 0 {
			return nil, ErrQuoteIndex
		}
	}

	// Get quote and send it
	return NewSimpleSend(msg.ChannelID, quo.List[ind]), nil
}

/* quote list */

type QuoteList struct{}

func NewQuoteList() *QuoteList { return &QuoteList{} }

func (q *QuoteList) Aliases() []string { return []string{"quote list", "quote ls"} }

func (q *QuoteList) Desc() string { return "Lists all approved quotes." }

func (q *QuoteList) Roles() []string { return nil }

func (q *QuoteList) Chans() []string { return nil }

func (q *QuoteList) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Get all approved quotes from db
	var quo quotes
	err := DBGet(&quotes{}, KeyQuotes, &quo)
	if err == ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// List them
	out := utils.Under("Quotes:") + "\n"
	for i, q := range quo.List {
		if len(q) > QuoteLineLimit {
			q = q[:QuoteLineLimit] + "[...]"
		}
		out += utils.Bold("#"+strconv.Itoa(i)+":") + " " + q + "\n"
	}

	return NewSimpleSend(msg.ChannelID, out), nil
}

/* quote pending */

type QuotePending struct{}

func NewQuotePending() *QuotePending { return &QuotePending{} }

func (q *QuotePending) Aliases() []string { return []string{"quote pending", "quote pd"} }

func (q *QuotePending) Desc() string { return "Lists all pending quotes." }

func (q *QuotePending) Roles() []string { return nil }

func (q *QuotePending) Chans() []string { return nil }

func (q *QuotePending) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Get all pending quotes from db
	var pen quotes
	err := DBGet(&quotes{}, KeyPending, &pen)
	if err == ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// List them
	out := utils.Under("Pending Quotes:") + "\n"
	for i, q := range pen.List {
		if len(q) > QuoteLineLimit {
			q = q[:QuoteLineLimit] + "[...]"
		}
		out += utils.Bold("#"+strconv.Itoa(i)+":") + " " + q + "\n"
	}

	return NewSimpleSend(msg.ChannelID, out), nil
}

/* quote add */

type QuoteAdd struct {
	Quote string `arg:"quote"`
}

func NewQuoteAdd() *QuoteAdd { return &QuoteAdd{} }

func (q *QuoteAdd) Aliases() []string { return []string{"quote add"} }

func (q *QuoteAdd) Desc() string { return "Adds a quote to the pending list." }

func (q *QuoteAdd) Roles() []string { return nil }

func (q *QuoteAdd) Chans() []string { return nil }

func (q *QuoteAdd) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Check args
	if len(args) == 0 {
		return nil, ErrQuoteArgs
	}

	// Get the pending quote list from the db
	var pen quotes
	err := DBGet(&quotes{}, KeyPending, &pen)
	if err == ErrDBNotFound {
		// Create a new quote list
		pen = quotes{
			List: []string{},
			Last: -1,
		}
	} else if err != nil {
		return nil, err
	}

	// Join args
	newQuote := strings.Join(args, " ")

	// Put the new quote into the pending quote list and update Last
	pen.List = append(pen.List, newQuote)
	pen.Last++

	// Set the pending quote list in the db
	_, _, err = DBSet(&pen, KeyPending)
	if err != nil {
		return nil, err
	}

	// Send message to channel
	out := "Added" + utils.Block(newQuote) + "to the Pending list at index "
	out += utils.Code(strconv.Itoa(pen.Last))
	return NewSimpleSend(msg.ChannelID, out), nil
}

/* quote approve */

type QuoteApprove struct {
	Index int `arg:"index"`
}

func NewQuoteApprove() *QuoteApprove { return &QuoteApprove{} }

func (q *QuoteApprove) Aliases() []string { return []string{"quote approve"} }

func (q *QuoteApprove) Desc() string { return "Approves a quote." }

func (q *QuoteApprove) Roles() []string { return []string{"mod"} }

func (q *QuoteApprove) Chans() []string { return nil }

func (q *QuoteApprove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Check args
	if len(args) == 0 {
		return nil, ErrQuoteArgs
	}

	// Get pending list
	var pen quotes
	err := DBGet(&quotes{}, KeyPending, &pen)
	if err == ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// Check index
	ind, err := strconv.Atoi(args[0])
	if err != nil || ind < 0 || ind > pen.Last {
		return nil, ErrQuoteIndex
	}

	// Get approved list
	var quo quotes
	err = DBGet(&quotes{}, KeyQuotes, &quo)
	if err == ErrDBNotFound {
		quo = quotes{
			List: []string{},
			Last: -1,
		}
	} else if err != nil {
		return nil, err
	}

	// Move pending quote to approved list, filling gaps first
	if quo.Last == -1 {
		quo.List = append(quo.List, pen.List[ind])
		quo.Last++
	} else {
		ins := quo.Last + 1
		for i, q := range quo.List {
			if len(q) == 0 {
				ins = i
				break
			}
		}
		if ins > quo.Last {
			quo.List = append(quo.List, pen.List[ind])
			quo.Last = ins
		} else {
			quo.List[ins] = pen.List[ind]
		}
	}

	// Reorder pending list
	newPen := pen.List[:ind]
	if ind != pen.Last {
		newPen = append(newPen, pen.List[ind+1:]...)
	}
	pen.List = newPen
	pen.Last--

	// Set quotes and pending
	_, _, err = DBSet(&pen, KeyPending)
	if err != nil {
		return nil, err
	}
	_, _, err = DBSet(&quo, KeyQuotes)
	if err != nil {
		return nil, err
	}

	out := "Approved quote\n" + utils.Block(quo.List[len(quo.List)-1]) + "now at index "
	out += utils.Code(strconv.Itoa(quo.Last))

	return NewSimpleSend(msg.ChannelID, out), nil
}

/* quote remove */

type QuoteRemove struct {
	Index int `arg:"index"`
}

func NewQuoteRemove() *QuoteRemove { return &QuoteRemove{} }

func (q *QuoteRemove) Aliases() []string { return []string{"quote remove", "quote rm"} }

func (q *QuoteRemove) Desc() string { return "Removes a quote." }

func (q *QuoteRemove) Roles() []string { return []string{"mod"} }

func (q *QuoteRemove) Chans() []string { return nil }

func (q *QuoteRemove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Check args
	if len(args) == 0 {
		return nil, ErrQuoteArgs
	}

	// Get quotes list
	var quo quotes
	err := DBGet(&quotes{}, KeyQuotes, &quo)
	if err == ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// Check index
	ind, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}
	if ind < 0 || ind > quo.Last {
		return nil, ErrQuoteIndex
	}

	// Clear index, don't reorder
	rem := quo.List[ind]
	quo.List[ind] = ""
	if ind == quo.Last {
		// Change last to first non-clear last
		for i := quo.Last; i >= 0; i-- {
			if len(quo.List[i]) > 0 {
				quo.Last = i
				break
			}
		}
	}

	// Set quotes
	_, _, err = DBSet(&quo, KeyQuotes)
	if err != nil {
		return nil, err
	}

	out := "Rejected quote\n" + utils.Block(rem)
	return NewSimpleSend(msg.ChannelID, out), nil
}

/* quote reject */

type QuoteReject struct {
	Index int `arg:"index"`
}

func NewQuoteReject() *QuoteReject { return &QuoteReject{} }

func (q *QuoteReject) Aliases() []string { return []string{"quote reject", "quote rj"} }

func (q *QuoteReject) Desc() string { return "Rejects a quote from the pending list." }

func (q *QuoteReject) Roles() []string { return nil }

func (q *QuoteReject) Chans() []string { return nil }

func (q *QuoteReject) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Check args
	if len(args) == 0 {
		return nil, ErrQuoteArgs
	}

	// Get pending list
	var pen quotes
	err := DBGet(&quotes{}, KeyPending, &pen)
	if err == ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// Check index
	ind, err := strconv.Atoi(args[0])
	if err != nil || ind < 0 || ind > pen.Last {
		return nil, ErrQuoteIndex
	}

	// Reorder list
	rem := pen.List[ind]
	newPen := pen.List[:ind]
	if ind != pen.Last {
		newPen = append(newPen, pen.List[ind+1:]...)
	}
	pen.List = newPen
	pen.Last--

	// Set pending
	_, _, err = DBSet(&pen, KeyPending)
	if err != nil {
		return nil, err
	}

	out := "Rejected quote\n" + utils.Block(rem)
	return NewSimpleSend(msg.ChannelID, out), nil
}
