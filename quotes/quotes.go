// Package quotes is your friendly neighbourhood quote monster. It can record, play back and delete quotes.
package quotes

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/doenietzomoeilijk/go-chat-bot-plugins/authorization"
	"github.com/go-chat-bot/bot"
	_ "github.com/mattn/go-sqlite3" // Use sqlite3
)

var (
	db      *sql.DB
	delStmt *sql.Stmt
)

// Quote holds a singular quote
type Quote struct {
	ID        uint      `db:"id"`
	Channel   string    `db:"channel"`
	Author    string    `db:"author"`
	Timestamp time.Time `db:"timestamp"`
	Content   string    `db:"content"`
}

// Set up our DB - make sure it exists and that it has the proper table.
func setupDb() {
	var err error

	db, err = sql.Open("sqlite3", "quotes.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS quotes (
        id INTEGER NOT NULL PRIMARY KEY,
        channel VARCHAR(50) NOT NULL DEFAULT 'unknown',
        author VARCHAR(15) NOT NULL DEFAULT 'unknown',
        timestamp DATETIME NULL,
        deleted TINYINT UNSIGNED NOT NULL DEFAULT 0,
        content TEXT
    )`)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS chan ON quotes(channel)`)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idel ON quotes(deleted)`)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS cont ON quotes(content)`)

	if delStmt, err = db.Prepare("UPDATE quotes SET deleted = 1 WHERE channel = ? and id = ?"); err != nil {
		log.Fatalln("Couldn't prepare statement:", err)
	}
}

// Add a quote to the database.
func addQuote(command *bot.Cmd) (msg string, err error) {
	var insertID int64

	author, err := authorization.Authorize(command.ChannelData, "author", command.User)
	if err != nil {
		log.Println("Could not authorize:", err)
		return "", nil
	}

	quote := strings.Trim(command.RawArgs, " ")
	if quote == "" {
		return
	}

	res, err := db.Exec(
		`INSERT INTO quotes
        (channel, author, timestamp, content)
        VALUES (?, ?, CURRENT_TIMESTAMP, ?)`,
		command.Channel,
		author,
		quote)

	if err != nil {
		log.Println("Could not insert quote:", err)
		err = errors.New("could not insert quote")
		return
	}

	insertID, err = res.LastInsertId()
	if err == nil {
		msg = fmt.Sprintf("Quote inserted with id %d.", insertID)
		log.Println(msg)
	} else {
		log.Println("error while inserting quote", err)
		err = nil
	}

	return
}

// Actually run the query, and get a specific result. If num > 0 it'll specify a specific result, if it's -1 it'll give you the latest quote.
func queryToQuote(w map[string]interface{}, num int) (Q Quote) {
	fields := "id, channel, author, timestamp, content"
	query := "SELECT %s FROM quotes WHERE deleted = 0"
	var binds []interface{}
	var usedLike bool
	var err error

	for where, bind := range w {
		query += fmt.Sprintf(" AND %s", where)
		binds = append(binds, bind)
		if strings.Contains(where, "LIKE") {
			usedLike = true
		}
	}

	if num == -1 {
		query += " ORDER BY id DESC LIMIT 1"
	}

	var count = 1
	if num != -1 {
		countRow := db.QueryRow(fmt.Sprintf(query, "COUNT(*)"), binds...)
		err = countRow.Scan(&count)

		if count == 0 || err != nil {
			return
		}
	}

	if count > 1 {
		if num == 0 {
			num = int(rand.Intn(count)) + 1
		}

		query += fmt.Sprintf(" LIMIT 1 OFFSET %d", num-1)
	}

	query = fmt.Sprintf(query, fields)

	row := db.QueryRow(query, binds...)
	err = row.Scan(&Q.ID, &Q.Channel, &Q.Author, &Q.Timestamp, &Q.Content)
	if count > 1 && usedLike {
		Q.Content += fmt.Sprintf(" (%d/%d)", num, count)
	}

	if err != nil {
		Q = Quote{}

	}

	return
}

// Fetch a quote from the database
func getQuote(command *bot.Cmd) (msg string, err error) {
	args := command.Args
	var lastArgNum int

	if len(command.Args) > 0 {
		lastArg := args[len(args)-1]
		lan, err := strconv.Atoi(lastArg)
		if err != nil {
			lastArgNum = int(0)
			err = nil
		} else {
			lastArgNum = int(lan)
			args = args[:len(args)-1]

		}
	}

	m := make(map[string]interface{})
	m["channel = ?"] = command.Channel

	switch {
	case len(args) == 1 && args[0] == "-id":
		m["id = ?"] = lastArgNum
	case len(args) > 0:
		m["content LIKE ?"] = fmt.Sprintf("%%%s%%", strings.Join(args, " "))
	}

	Q := queryToQuote(m, lastArgNum)
	if Q.ID != 0 {
		msg = fmt.Sprintf("#%d: %s", Q.ID, Q.Content)
	}

	return
}

// Fetch the latest quote from the database.
func getLastQuote(command *bot.Cmd) (msg string, err error) {
	m := make(map[string]interface{})
	m["channel = ?"] = command.Channel
	Q := queryToQuote(m, -1)
	if Q.ID != 0 {
		msg = fmt.Sprintf("#%d: %s", Q.ID, Q.Content)
	}

	return
}

// Delete a quote from the database.
func deleteQuote(c *bot.Cmd) (msg string, err error) {
	if len(c.Args) < 2 {
		return "Specify which quote you want to delete with `-id <id>`.", nil
	}

	author, err := authorization.Authorize(c.ChannelData, "author", c.User)
	if err != nil {
		log.Println("Could not authorize:", err)
		return "", nil
	}

	id := 0
	if len(c.Args) == 2 && c.Args[0] == "-id" {
		id, _ = strconv.Atoi(c.Args[1])
	}

	if id == 0 {
		return
	}

	res, _ := delStmt.Exec(c.Channel, id)
	affected, _ := res.RowsAffected()
	if affected == 0 {
		msg = "Could not delete quote."
	} else {
		log.Printf("Quote id %d deleted by %s.", id, author)
		msg = "Quote deleted."
	}

	return
}

// Fetch info about a quote and display it in the channel.
func quoteInfo(c *bot.Cmd) (msg string, err error) {
	if len(c.Args) < 2 {
		return "Specify which quote you want to show info for with `-id <id>`.", nil
	}

	id := 0
	if len(c.Args) == 2 && c.Args[0] == "-id" {
		id, _ = strconv.Atoi(c.Args[1])
	}

	if id == 0 {
		return
	}

	m := make(map[string]interface{})
	m["channel = ?"] = c.Channel
	m["id = ?"] = id
	Q := queryToQuote(m, -1)
	if Q.ID != 0 {
		msg = fmt.Sprintf("#%d: quoted by %s on %s", Q.ID, Q.Author, Q.Timestamp.Format("2006-01-02, 15:04"))
	}

	return
}

func init() {
	setupDb()

	bot.RegisterCommand(
		"addquote",
		"Add a quote to the bot",
		"",
		addQuote)

	bot.RegisterCommand(
		"quote",
		"Get a quote from the bot",
		"[-id <id>|<[querystring] [<nth match>]]",
		getQuote)

	bot.RegisterCommand(
		"q",
		"Alias for !quote",
		"",
		getQuote)

	bot.RegisterCommand(
		"lastquote",
		"Get the last quote that was added in this channel",
		"",
		getLastQuote)
	bot.RegisterCommand(
		"delquote",
		"Delete a quote",
		"-id <id>",
		deleteQuote)
	bot.RegisterCommand(
		"quoteinfo",
		"Get info about a quote",
		"-id <id>",
		quoteInfo)
}
