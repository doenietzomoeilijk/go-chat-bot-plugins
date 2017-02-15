package echo

import (
	"fmt"

	"github.com/doenietzomoeilijk/go-chat-bot-plugins/authorization"
	"github.com/go-chat-bot/bot"
)

func init() {
	bot.RegisterCommand(
		"echo",
		"",
		"",
		func(command *bot.Cmd) (msg string, err error) {
			msg = fmt.Sprintf(
				"channel=%s, hostmask=%s, args=%#v",
				command.Channel,
				authorization.FullHostmask(command.User),
				command.Args)

			return
		})
}
