package handlers

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sahilm/fuzzy"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/internal/utils"
)

const (
	keyPending = "pending"
	keyQuotes  = "approve"

	quoteLineLimit = 50
)

var (
	// ErrQuoteIndex means quote index is not valid
	ErrQuoteIndex = errors.New("index not valid")
	// ErrQuoteEmpty means quote list is not there
	ErrQuoteEmpty = errors.New("quote list not initialised")
	// ErrQuoteNone means user entered no quote
	ErrQuoteNone = errors.New("no quote entered, please enter a quote")
	// ErrQueryNone means user entered no quote
	ErrQueryNone = errors.New("no search terms entered")
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

type quote struct {
	nilCommand
	Index []int `arg:"index"`
}

func newQuote() *quote { return &quote{} }

func (q *quote) Aliases() []string { return []string{"quote"} }

func (q *quote) Desc() string { return "Get a quote at given index. No args gives a random quote." }

func (q *quote) Subcommands() []commands.Command {
	return []commands.Command{
		newQuoteAdd(),
		newQuoteApprove(),
		newQuoteList(),
		newQuotePending(),
		newQuoteRemove(),
		newQuoteReject(),
		newQuoteSearch(),
	}
}

func (q *quote) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// Get quotes
	var quo quotes
	err := commands.DBGet(&quotes{}, keyQuotes, &quo)
	if err == commands.ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// Check args
	var ind int
	if len(q.Index) == 0 {
		// Gen random number
		rand.Seed(time.Now().UnixNano())
		ind = rand.Intn(len(quo.List))
	} else {
		ind = q.Index[0]
		if ind > quo.Last || ind < 0 {
			return nil, ErrQuoteIndex
		}
	}

	// Get quote and send it
	return commands.NewSimpleSend(msg.ChannelID, quo.List[ind]), nil
}

type quoteAdd struct {
	nilCommand
	New []string `arg:"quote"`
}

func newQuoteAdd() *quoteAdd { return &quoteAdd{} }

func (q *quoteAdd) Aliases() []string { return []string{"quote add"} }

func (q *quoteAdd) Desc() string { return "Adds a quote to the pending list." }

func (q *quoteAdd) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// Get the pending quote list from the db
	var pen quotes
	err := commands.DBGet(&quotes{}, keyPending, &pen)
	if err == commands.ErrDBNotFound {
		// Create a new quote list
		pen = quotes{
			List: []string{},
			Last: -1,
		}
	} else if err != nil {
		return nil, err
	}

	// Check quote first
	newQuote := strings.TrimSpace(strings.Join(q.New, " "))

	if len(newQuote) == 0 {
		// Quote is empty, throw error
		return nil, ErrQuoteNone
	}

	// Put the new quote into the pending quote list and update Last
	newQuote = strings.Join(q.New, " ")

	pen.List = append(pen.List, newQuote)
	pen.Last++

	// Set the pending quote list in the db
	_, _, err = commands.DBSet(&pen, keyPending)
	if err != nil {
		return nil, err
	}

	// Send message to channel
	out := "Added" + utils.Block(newQuote) + "to the Pending list at index "
	out += utils.Code(strconv.Itoa(pen.Last))
	return commands.NewSimpleSend(msg.ChannelID, out), nil
}

type quoteApprove struct {
	nilCommand
	Index int `arg:"index"`
}

func newQuoteApprove() *quoteApprove { return &quoteApprove{} }

func (q *quoteApprove) Aliases() []string { return []string{"quote approve", "quote ap"} }

func (q *quoteApprove) Desc() string { return "Approves a quote." }

func (q *quoteApprove) Roles() []string { return []string{"mod"} }

func (q *quoteApprove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// Get pending list
	var pen quotes
	err := commands.DBGet(&quotes{}, keyPending, &pen)
	if err == commands.ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// Check index
	if err != nil || q.Index < 0 || q.Index > pen.Last {
		return nil, ErrQuoteIndex
	}

	// Get approved list
	var quo quotes
	err = commands.DBGet(&quotes{}, keyQuotes, &quo)
	if err == commands.ErrDBNotFound {
		quo = quotes{
			List: []string{},
			Last: -1,
		}
	} else if err != nil {
		return nil, err
	}

	// Move pending quote to approved list, filling gaps first
	var ins int
	if quo.Last == -1 {
		// quote list is empty
		quo.List = append(quo.List, pen.List[q.Index])
		quo.Last++
		ins = quo.Last
	} else {
		// quote list is not empty
		ins = quo.Last + 1
		for i, q := range quo.List {
			if len(q) == 0 {
				ins = i
				break
			}
		}
		if ins > quo.Last {
			quo.List = append(quo.List, pen.List[q.Index])
			quo.Last = ins
		} else {
			quo.List[ins] = pen.List[q.Index]
		}
	}

	// Reorder pending list
	newPen := pen.List[:q.Index]
	if q.Index != pen.Last {
		newPen = append(newPen, pen.List[q.Index+1:]...)
	}
	pen.List = newPen
	pen.Last--

	// Set quotes and pending
	_, _, err = commands.DBSet(&pen, keyPending)
	if err != nil {
		return nil, err
	}
	_, _, err = commands.DBSet(&quo, keyQuotes)
	if err != nil {
		return nil, err
	}

	out := "Approved quote\n" + utils.Block(quo.List[ins]) + "now at index "
	out += utils.Code(strconv.Itoa(ins))

	return commands.NewSimpleSend(msg.ChannelID, out), nil
}

type quoteList struct {
	nilCommand
}

func newQuoteList() *quoteList { return &quoteList{} }

func (q *quoteList) Aliases() []string { return []string{"quote list", "quote ls"} }

func (q *quoteList) Desc() string { return "Lists all approved quotes." }

func (q *quoteList) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// Get all approved quotes from db
	var quo quotes
	err := commands.DBGet(&quotes{}, keyQuotes, &quo)
	snd := commands.NewSend(msg.ChannelID)

	if err == commands.ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// List them
	var tmp string
	var out = utils.Under("Quotes:") + "\n"
	for i, q := range quo.List {
		// deleted quote, skip
		if len(q) == 0 {
			continue
		}

		if len(q) > quoteLineLimit {
			q = q[:quoteLineLimit] + "[...]"
		}

		tmp = utils.Bold("#"+strconv.Itoa(i)+":") + " " + q + "\n"

		if len(out)+len(tmp) >= commands.MessageLimit {
			// reached message limit, add a new message to Send
			snd.Message(out)
			out = tmp
		} else {
			out += tmp
		}
	}

	return snd.Message(out), nil
}

type quotePending struct {
	nilCommand
}

func newQuotePending() *quotePending { return &quotePending{} }

func (q *quotePending) Aliases() []string { return []string{"quote pending", "quote pd"} }

func (q *quotePending) Desc() string { return "Lists all pending quotes." }

func (q *quotePending) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// Get all pending quotes from db
	var pen quotes
	err := commands.DBGet(&quotes{}, keyPending, &pen)
	if err == commands.ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// Check empty
	if len(pen.List) == 0 {
		return commands.NewSimpleSend(msg.ChannelID, "Pending list is empty."), nil
	}

	// List them
	out := utils.Under("Pending quotes:") + "\n"
	for i, q := range pen.List {
		if len(q) > quoteLineLimit {
			q = q[:quoteLineLimit] + "[...]"
		}
		out += utils.Bold("#"+strconv.Itoa(i)+":") + " " + q + "\n"
	}

	return commands.NewSimpleSend(msg.ChannelID, out), nil
}

type quoteReject struct {
	nilCommand
	Index int `arg:"index"`
}

func newQuoteReject() *quoteReject { return &quoteReject{} }

func (q *quoteReject) Aliases() []string { return []string{"quote reject", "quote rj"} }

func (q *quoteReject) Desc() string { return "Rejects a quote from the pending list." }

func (q *quoteReject) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// Get pending list
	var pen quotes
	err := commands.DBGet(&quotes{}, keyPending, &pen)
	if err == commands.ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// Check index
	if q.Index < 0 || q.Index > pen.Last {
		return nil, ErrQuoteIndex
	}

	// Reorder list
	rej := pen.List[q.Index]
	newPen := pen.List[:q.Index]
	if q.Index != pen.Last {
		newPen = append(newPen, pen.List[q.Index+1:]...)
	}
	pen.List = newPen
	pen.Last--

	// Set pending
	_, _, err = commands.DBSet(&pen, keyPending)
	if err != nil {
		return nil, err
	}

	out := "Rejected quote\n" + utils.Block(rej)
	return commands.NewSimpleSend(msg.ChannelID, out), nil
}

type quoteRemove struct {
	nilCommand
	Index int `arg:"index"`
}

func newQuoteRemove() *quoteRemove { return &quoteRemove{} }

func (q *quoteRemove) Aliases() []string { return []string{"quote remove", "quote rm"} }

func (q *quoteRemove) Desc() string { return "Removes a quote." }

func (q *quoteRemove) Roles() []string { return []string{"mod"} }

func (q *quoteRemove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// Get quotes list
	var quo quotes
	err := commands.DBGet(&quotes{}, keyQuotes, &quo)
	if err == commands.ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// Check index
	if q.Index < 0 || q.Index > quo.Last {
		return nil, ErrQuoteIndex
	}

	// Clear index, don't reorder
	rem := quo.List[q.Index]
	quo.List[q.Index] = ""
	if q.Index == quo.Last {
		// Change last to first non-clear last
		for i := quo.Last; i >= 0; i-- {
			if len(quo.List[i]) > 0 {
				quo.Last = i
				break
			}
		}
	}

	// Set quotes
	_, _, err = commands.DBSet(&quo, keyQuotes)
	if err != nil {
		return nil, err
	}

	out := "Removed quote\n" + utils.Block(rem)
	return commands.NewSimpleSend(msg.ChannelID, out), nil
}

type quoteSearch struct {
	nilCommand
	Query []string `arg:"query"`
}

func newQuoteSearch() *quoteSearch { return &quoteSearch{} }

func (q *quoteSearch) Aliases() []string { return []string{"quote search", "quote se"} }

func (q *quoteSearch) Desc() string { return "Searches for a quote, returns top 5 results." }

func (q *quoteSearch) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	// Join query
	qry := strings.TrimSpace(strings.Join(q.Query, " "))
	if len(qry) == 0 {
		return nil, ErrQueryNone
	}

	// Get quotes list
	var quo quotes
	err := commands.DBGet(&quotes{}, keyQuotes, &quo)
	if err == commands.ErrDBNotFound {
		return nil, ErrQuoteEmpty
	} else if err != nil {
		return nil, err
	}

	// use fuzzy finding to get top 5 results
	mat := fuzzy.Find(qry, quo.List)
	if len(mat) == 0 {
		return commands.NewSimpleSend(msg.ChannelID, "No matches found."), nil
	}

	// print results
	out := "Search Results:\n"
	for i, m := range mat {
		if i == 5 {
			break
		}
		out += utils.Bold("#"+strconv.Itoa(m.Index)+": ") + m.Str + "\n"
	}

	return commands.NewSimpleSend(msg.ChannelID, out), nil
}
