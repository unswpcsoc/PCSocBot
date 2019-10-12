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
	ErrNoInput        = errors.New("No Course code entered")
	ErrInvalidFormat  = errors.New("Invalid Course Code Format")
	ErrNotFound       = errors.New("Course Not Found")
	ErrScrapingFailed = errors.New("Web scraping failed, please contact an Exec")
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
		lineString := string(line)
		if title == "" {
			if strings.Contains(lineString, "module-title") {
				titleSlice := strings.Split(lineString, ">")
				if len(titleSlice) >= 3 {
					title = strings.Split(titleSlice[2], "<")[0]
				} else {
					return nil, ErrScrapingFailed
				}
			}

		}
		if desc == "" {
			if strings.Contains(lineString, "readmore__wrapper") {
				if len(lines) <= i+2 {
					return nil, ErrScrapingFailed
				}
				desc = string(lines[i+2])
				desc = html.UnescapeString(desc)
				// if the desc have html formatting (<>) before it, take the first entry
				if string(desc[0]) == "<" {
					descSlice := strings.Split(desc, ">")
					if len(descSlice) >= 2 {
						desc = strings.Split(descSlice[1], "<")[0]
					} else {
						return nil, ErrScrapingFailed
					}
				}
			}
		}
		if term == "" {
			if strings.Contains(lineString, ">Offering Terms<") {
				if len(lines) <= i+1 {
					return nil, ErrScrapingFailed
				}
				termSlice := strings.Split(string(lines[i+1]), ">")
				if len(termSlice) >= 2 {
					term = strings.Split(termSlice[1], "<")[0]
				} else {
					return nil, ErrScrapingFailed
				}
			}
		}
		if cond == "" {
			if strings.Contains(lineString, "Prerequisite") {
				condSlice := strings.Split(lineString, ">")
				if len(condSlice) >= 2 {
					cond = strings.Split(condSlice[1], "<")[0]
				} else {
					return nil, ErrScrapingFailed
				}
			}
		}
		if title != "" && desc != "" && term != "" && cond != "" {
			// stop early if all fields are filled
			break
		}
	}
	// If no prerequisites/offering found
	if cond == "" {
		cond = "None"
	}
	if term == "" {
		term = "None"
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
