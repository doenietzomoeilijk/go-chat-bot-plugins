# Quotes

On startup, it will create an sqlite3 database `quotes.db` in your bot's working directory.

Depending on your user setup (see the documentation about `authorization`), users now have the following commands at their disposal:

* `!addquote <your quote>` Add a quote to the database.

* `!quote` Without any parameters, returns a random quote.

* `!quote <words> [<num>]` Add words to match quotes against them. If several quotes match, an optional `<num>` requests a specific one.

* `!quote -id <id>` All quotes have an ID; use this syntax to request a specific quote.