# go-chat-bot-plugins

A collection of plugins for [go-chat-bot](https://github.com/go-chat-bot/bot).

* **authorization**: authorize users against a userfile
* **quotes**: add, display and remove quotes for a channel

## authorization

Expects a file `userfile.json` to be present in your bot's working directory.
See `authorization/userfile_dist.json` for a skeleton file.

## quotes

Will create an sqlite3 database `quotes.db` in your bot's working directory.