// Package substitute keeps track of what people have said, allowing you to semi-regex-correct them using s/word/new word/. This doesn't use regex syntax, just a simple replace().
package substitute

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-chat-bot/bot"
)

type nickAndLine struct {
	nick string
	line string
}

var (
	lines     map[string][]nickAndLine
	re        *regexp.Regexp
	keepLines int
	bold      string
)

func handleSub(cmd *bot.PassiveCmd) (msg string, err error) {
	linesLen := len(lines[cmd.Channel])
	if found := re.FindStringSubmatch(cmd.Raw); found != nil && found[1] != " " {
		// Found the string, now loop through our history and see if we can do
		// a substitution.
		// found[2] = strings.Replace(bold, "", found[2], -1)
		for i := linesLen; i > 0; i-- {
			l := lines[cmd.Channel][i-1]
			if strings.Contains(l.line, found[1]) {
				repl := fmt.Sprintf("%s%s%s", bold, found[2], bold)
				replaced := strings.Replace(l.line, found[1], repl, -1)
				msg = fmt.Sprintf("<%s> %s", l.nick, replaced)

				return
			}
		}
	} else {
		// Not found, just add this line (making sure we don't cross our limit)
		if linesLen >= keepLines {
			lines[cmd.Channel] = lines[cmd.Channel][1:]
		}

		lines[cmd.Channel] = append(lines[cmd.Channel], nickAndLine{cmd.User.Nick, cmd.Raw})
	}

	return
}

func init() {
	keepLines = 20
	bold = fmt.Sprintf("%c", 2)
	re = regexp.MustCompile(`^s/([^\/]+)/([^\/]+)/?`)
	lines = make(map[string][]nickAndLine)
	bot.RegisterPassiveCommand("substitute", handleSub)
}
