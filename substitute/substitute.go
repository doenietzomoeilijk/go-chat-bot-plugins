// Package substitute keeps track of what people have said, allowing you to semi-regex-correct them using s/word/new word/. This doesn't use regex syntax, just a simple replace().
package substitute

import (
	"regexp"

	"fmt"

	"strings"

	"github.com/go-chat-bot/bot"
)

type nickAndLine struct {
	nick string
	line string
}

var (
	lines     []nickAndLine
	re        *regexp.Regexp
	keepLines int
)

func handleSub(cmd *bot.PassiveCmd) (msg string, err error) {
	linesLen := len(lines)
	if found := re.FindStringSubmatch(cmd.Raw); found != nil {
		// Found the string, now loop through our history and see if we can do
		// a substitution.
		for i := linesLen; i > 0; i-- {
			l := lines[i-1]
			if strings.Contains(l.line, found[1]) {
				repl := fmt.Sprintf("%c%s%c", 2, found[2], 2)
				replaced := strings.Replace(l.line, found[1], repl, -1)
				msg = fmt.Sprintf("<%s> %s", l.nick, replaced)
			}
		}
	} else {
		// Not found, just add this line (making sure we don't cross our limit)
		if linesLen >= keepLines {
			lines = lines[1:]
		}

		lines = append(lines, nickAndLine{cmd.User.Nick, cmd.Raw})
	}

	return
}

func init() {
	keepLines = 20
	re = regexp.MustCompile(`^s/([^\/]+)/([^\/]+)/?`)
	bot.RegisterPassiveCommand("substitute", handleSub)
}
