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
func (ri *RandomItem) GetResponse() string {
	length := len(ri.Responses)
	if length < 1 {
		log.Println("this RandomItem has no responses...")
		return ""
	}

	rnd := rand.Intn(length - 1)
	log.Printf("Grabbing random item %d/%d from %s", rnd, length, ri.Keyword)

	return ri.Responses[rnd]
}

var (
	// Randomlines holds our RandomItems, mapped to their primary keyword.
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
		log.Fatalln("Couldn't read file:", err)
	}

	var parts []RandomItem
	json.Unmarshal(file, &parts)

	for _, part := range parts {
		bareKeyword := strings.TrimLeft(part.Keyword, bot.CmdPrefix)
		isPassive := bareKeyword == part.Keyword
		Randomlines[bareKeyword] = part
		maybeAdd(bareKeyword, isPassive)

		for _, alias := range part.Aliases {
			bareAlias := strings.TrimLeft(alias, bot.CmdPrefix)
			isPassiveAlias := bareAlias == alias
			aliasToKeyword[bareAlias] = bareKeyword
			maybeAdd(bareAlias, isPassiveAlias)
		}
	}

	// Set up the regex for the passive command.
	pattern = fmt.Sprintf("(?i)\\b(%s)\\b", strings.Join(passives, "|"))
	re = regexp.MustCompile(pattern)
}

func maybeAdd(s string, p bool) {
	if p {
		passives = append(passives, s)
	} else {
		actives = append(actives, s)
		addCommand(s)
	}
}

func addCommand(s string) {
	if _, ok := Randomlines[s]; ok {
		bot.RegisterCommand(
			s,
			fmt.Sprintf("Random line for %s", s),
			"",
			func(c *bot.Cmd) (msg string, err error) {
				rnd, _ := Randomlines[s]
				msg = rnd.GetResponse()
				return
			},
		)
	}
}

func randompassiveline(c *bot.PassiveCmd) (msg string, err error) {
	if found := re.FindString(c.Raw); found != "" {
		keyword := aliasToKeyword[found]
		rnd, _ := Randomlines[keyword]
		msg = rnd.GetResponse()
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

			return
		},
	)
	bot.RegisterPassiveCommand("randomline", randompassiveline)
}
