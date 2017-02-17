package randomline

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"

	"regexp"

	"strings"

	"fmt"

	"github.com/go-chat-bot/bot"
)

// RandomItem holds one item stanza.
type RandomItem struct {
	Keyword   string
	Aliases   []string
	Responses []string
}

// GetResponse gives you one of the RandomItem's Responses.
func (ri RandomItem) GetResponse() string {
	length := len(ri.Responses)
	rnd := rand.Intn(length - 1)
	log.Printf("Grabbing random item %d/%d from %s", rnd, length, ri.Keyword)

	return ri.Responses[rnd]
}

func (ri RandomItem) Dump() string {
	return fmt.Sprintf("keyword=%s aliases=%v responses=%#v", ri.Keyword, ri.Aliases, ri.Responses)
}

// Randomlines holds our RandomItems, mapped to their primary keyword.
var (
	Randomlines    map[string]RandomItem
	aliasToKeyword map[string]string
	passives       []string
	actives        []string
	pattern        string
	re             *regexp.Regexp
)

func loadRandomlines() {
	Randomlines = map[string]RandomItem{}
	passives = []string{}
	actives = []string{}
	pattern = ""
	aliasToKeyword = map[string]string{}

	file, err := ioutil.ReadFile("randomlines.json")
	if err != nil {
		log.Fatalf("Couldn't read file: %#v", err)
	}

	var parts []RandomItem
	json.Unmarshal(file, &parts)

	for _, part := range parts {
		Randomlines[part.Keyword] = part

		maybeAdd(part.Keyword)
		for _, alias := range part.Aliases {
			maybeAdd(alias)
			aliasToKeyword[alias] = part.Keyword

		}
	}

	pattern = fmt.Sprintf("(?i)\\b(%s)\\b", strings.Join(passives, "|"))
	re = regexp.MustCompile(pattern)

	for _, random := range actives {
		bot.RegisterCommand(
			random,
			fmt.Sprintf("Random line for %s", random),
			"",
			func(c *bot.Cmd) (msg string, err error) {
				msg = Randomlines[random].GetResponse()
				return
			},
		)
	}
}

func maybeAdd(s string) {
	if strings.HasPrefix(s, bot.CmdPrefix) {
		actives = append(actives, strings.TrimLeft(s, bot.CmdPrefix))
	} else {
		passives = append(passives, s)
	}
}

func randompassiveline(c *bot.PassiveCmd) (msg string, err error) {
	if found := re.FindString(c.Raw); found != "" {
		keyword := aliasToKeyword[found]
		msg = Randomlines[keyword].GetResponse()
	}

	return
}

func init() {
	loadRandomlines()

	bot.RegisterCommand(
		"loadrandomlines",
		"Reload the random lines file",
		"",
		func(c *bot.Cmd) (msg string, err error) {
			loadRandomlines()
			msg = "Loaded lines"

			return
		},
	)
	bot.RegisterCommand(
		"dump",
		"Dump the randomlines for a keyword",
		"keyword",
		func(c *bot.Cmd) (msg string, err error) {
			ri, ok := Randomlines[c.Args[0]]
			if !ok {
				msg = "not found"
			} else {
				ri.Dump()
			}

			return
		},
	)
	bot.RegisterPassiveCommand("randomline", randompassiveline)
}
