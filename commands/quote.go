package commands

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
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

/* helpers */

const (
	/* keys for quotes */
	KeyPending = "pending"
	KeyApprove = "approve"

	/* limits */
	QuoteListLimit = 50
)

// quotes implements the Storer interface
type quotes struct {
	List []string
	Last int
}

func (q *quotes) Index() string {
	return "quotes"
}

func (q *quote) Unmarshal(mar string) (Storer, error) {
	var res quotes
	err := json.Unmarshal([]byte(mar), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

/* quote */

type Quote struct {
	names []string
	desc  string
}

func NewQuote() *Quote {
	return &Quote{
		names: []string{"quote"},
		desc:  "Quote!",
	}
}

func (q *Quote) Names() []string { return q.names }

func (q *Quote) Desc() string { return q.desc }

func (q *Quote) Roles() []string { return nil }

func (q *Quote) Chans() []string { return nil }

func (q *Quote) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	return NewSimpleSend(msg.ChannelID, "Pong!"), nil
}

/* quote list */

type QuoteList struct {
	names []string
	desc  string
}

func NewQuoteList() *QuoteList {
	return &QuoteList{
		names: []string{"quote list", "quote ls"},
		desc:  "Lists all approved quotes.",
	}
}

func (q *QuoteList) Names() []string { return q.names }

func (q *QuoteList) Desc() string { return q.desc }

func (q *QuoteList) Roles() []string { return nil }

func (q *QuoteList) Chans() []string { return nil }

func (q *QuoteList) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Get all approved quotes from db
	quoteList, err := DBGet(&quotes{}, KeyApprove)
	if err != nil {
		return nil, err
	}

	// List them
	out := ""
	for i, q := range quoteList.List {
		if len(q) > QuoteListLimit {
			q = q[:QuoteListLimit]
		}
		out += utils.Bold("#"+string(i)+":") + " " + q + "\n"
	}

	return NewSimpleSend(msg.ChannelID, utils.Block(out))
}

/* quote pending */

type QuotePending struct {
	names []string
	desc  string
}

func NewQuotePending() *QuotePending {
	return &QuotePending{
		names: []string{"quote pending", "quote pd"},
		desc:  "Lists all pending quotes.",
	}
}

func (q *QuotePending) Names() []string { return q.names }

func (q *QuotePending) Desc() string { return q.desc }

func (q *QuotePending) Roles() []string { return nil }

func (q *QuotePending) Chans() []string { return nil }

func (q *QuotePending) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Get all pending quotes from db
	quoteList, err := DBGet(&quotes{}, KeyPending)
	if err != nil {
		return nil, err
	}

	// List them
	out := ""
	for _, q := range quoteList.List {
		if len(q) > QuoteListLimit {
			q = q[:QuoteListLimit]
		}
		out += utils.Bold("#"+string(i)+":") + " " + q + "\n"
	}

	return NewSimpleSend(msg.ChannelID, utils.Block(out))
}

/* quote approve */

type QuoteApprove struct {
	names []string
	desc  string
}

func NewQuoteApprove() *QuoteApprove {
	return &QuoteApprove{
		names: []string{"quote approve"},
		desc:  "Approves a quote to the pending list.",
	}
}

func (q *QuoteApprove) Names() []string { return q.names }

func (q *QuoteApprove) Desc() string { return q.desc }

func (q *QuoteApprove) Roles() []string { return nil }

func (q *QuoteApprove) Chans() []string { return nil }

func (q *QuoteApprove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	// Check args
	if len(args) == 0 {
		return nil, errors.New("not enough arguments")
	}

	// Get the pending quote list from the db
	var quoteList quotes
	var err error
	quoteList, err = DBGet(&quotes{}, KeyApprove)
	if err == ErrDBNotFound {
		// Create a new quote list
		quoteList = &quotes{
			List: []string{},
			Last: 0,
		}
	} else if err != nil {
		return nil, err
	}

	// Join args
	newQuote := strings.Join(args, " ")

	// Put the new quote into the pending quote list and update Last
	quoteList.List = append(quoteList.List, newQuote)
	quoteList.Last += 1

	// Set the pending quote list in the db
	_, _, err = DBSet(&quotes{}, quoteList)
	if err != nil {
		return nil, err
	}

	// Send message to channel
	out = "Approveed" + utils.Block(newQuote) + "to the Pending list at index " + utils.Code(quoteList.Last-1)
	return NewSimpleSend(msg.ChannelID, out)
}

/* quote remove */

type QuoteRemove struct {
	names []string
	desc  string
}

func NewQuoteRemove() *QuoteRemove {
	return &QuoteRemove{
		names: []string{"quote remove", "quote rm"},
		desc:  "Removes a quote from the pending list.",
	}
}

func (q *QuoteRemove) Names() []string { return q.names }

func (q *QuoteRemove) Desc() string { return q.desc }

func (q *QuoteRemove) Roles() []string { return nil }

func (q *QuoteRemove) Chans() []string { return nil }

func (q *QuoteRemove) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	return NewSimpleSend(msg.ChannelID, "Pong!"), nil
}

/* quote reject */

type QuoteReject struct {
	names []string
	desc  string
}

func NewQuoteReject() *QuoteReject {
	return &QuoteReject{
		names: []string{"quote reject", "quote rj"},
		desc:  "Rejects a quote from the pending list.",
	}
}

func (q *QuoteReject) Names() []string { return q.names }

func (q *QuoteReject) Desc() string { return q.desc }

func (q *QuoteReject) Roles() []string { return nil }

func (q *QuoteReject) Chans() []string { return nil }

func (q *QuoteReject) MsgHandle(ses *discordgo.Session, msg *discordgo.Message, args []string) (*CommandSend, error) {
	return NewSimpleSend(msg.ChannelID, "Pong!"), nil
}
