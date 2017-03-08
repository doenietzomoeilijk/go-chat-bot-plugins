// Package echo gives you a simple (but annoying) way to debug raw messages.
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
				"command, hostmask=%s, args=%#v, channel=%#v",
				authorization.Fullhost(command.User),
				command.Args,
				command.ChannelData,
			)

			return
		})
	bot.RegisterPassiveCommand(
		"echo",
		func(command *bot.PassiveCmd) (msg string, err error) {
			msg = fmt.Sprintf(
				"passivecommand, user=%+v, raw=%#v, channel=%#v",
				authorization.Fullhost(command.User),
				command.Raw,
				command.ChannelData,
			)

			return
		})
}
