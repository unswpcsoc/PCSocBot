package handlers

import (
	"bytes"
	"errors"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/unswpcsoc/PCSocBot/commands"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrNoInput       = errors.New("No Course code entered")
	ErrInvalidFormat = errors.New("Invalid Course Code Format")
	ErrNotFound      = errors.New("Course Not Found")
)

type handbook struct {
	nilCommand
	Code string `arg:"code"`
}

func newHandbook() *handbook { return &handbook{} }

func (h *handbook) Aliases() []string { return []string{"handbook"} }

func (h *handbook) Desc() string { return "Searches handbook.unsw for course" }

func (h *handbook) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {

	// Special case for DELL1234
	if strings.ToUpper(h.Code) == "DELL1234" {
		message := commands.NewSend(msg.ChannelID)
		embed := makeMessage("https://webapps.cse.unsw.edu.au/webcms2/course/index.php?cid=1137",
			"How to blow up my computer?",
			"No Course Outline (yet)",
			"None",
			"None")
		return message.Embed(embed), nil
	}

	// Check course code
	match, err := regexp.MatchString(`^[A-Za-z]{4}[0-9]{4}$`, h.Code)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, ErrInvalidFormat
	}

	//assume undergraduate
	url, htmlText, status, err := getResponse(h.Code, "undergraduate")
	if err != nil {
		return nil, err
	}

	//if not undergraduate then must be postgraduate
	if status != 200 {
		url, htmlText, status, err = getResponse(h.Code, "postgraduate")
	}
	if err != nil {
		return nil, err
	}

	if status != 200 {
		// probably a 404 error but whatever it is, the page could not be accessed
		return nil, ErrNotFound
	}

	// initialise info
	var title, desc, term, cond = "", "", "", ""

	// extract info via shifty means
	lines := bytes.Split(htmlText, []byte("\n"))
	for i, line := range lines {
		line := string(line[:])
		if title == "" {
			if strings.Contains(line, "module-title") {
				title = strings.Split(strings.Split(line, ">")[2], "<")[0]
			}

		}
		if desc == "" {
			if strings.Contains(line, "readmore__wrapper") {
				desc = strings.Split(strings.Split(string(lines[i+2][:]), ">")[1], "<")[0]
				desc = html.UnescapeString(desc)
			}
		}
		if term == "" {
			if strings.Contains(line, "<p tabindex=\"0\" class=\"\">") {
				term = strings.Split(strings.Split(line, ">")[1], "<")[0]
			}
		}
		if cond == "" {
			if strings.Contains(line, "Prerequisite") {
				cond = strings.Split(strings.Split(line, ">")[1], "<")[0]
			}
		}
		if title != "" && desc != "" && term != "" && cond != "" {
			// stop early if all fields are filled
			break
		}
	}
	// If no prerequisites found
	if cond == "" {
		cond = "None"
	}

	// create and send message
	message := commands.NewSend(msg.ChannelID)
	embed := makeMessage(url, title, desc, term, cond)
	return message.Embed(embed), nil
}

func getResponse(Code string, Graduate string) (url string, htmlText []byte, respCode int, err error) {
	// access page
	url = "https://www.handbook.unsw.edu.au/" + Graduate + "/courses/2020/" + Code
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, 0, err
	}
	// always close
	defer resp.Body.Close()
	htmlText, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, 0, err
	}

	return url, htmlText, resp.StatusCode, nil
}

func makeMessage(Url string, Title string, Desc string, Term string, Cond string) *discordgo.MessageEmbed {
	// create fields for embed
	terms := discordgo.MessageEmbedField{
		Name:   "Offering Terms",
		Value:  Term,
		Inline: true,
	}
	conds := discordgo.MessageEmbedField{
		Name:   "Enrolment Conditions",
		Value:  Cond,
		Inline: true,
	}
	messagefields := []*discordgo.MessageEmbedField{&terms, &conds}
	// make embed and return
	embed := &discordgo.MessageEmbed{
		URL:         Url,
		Title:       Title,
		Description: Desc,
		Fields:      messagefields,
		Color:       0xFDD600,
	}
	return embed
}
