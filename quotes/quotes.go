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
	db *sql.DB
)

// Quote holds a singular quote
type Quote struct {
	ID        uint      `db:"id"`
	Channel   string    `db:"channel"`
	Author    string    `db:"author"`
	Timestamp time.Time `db:"timestamp"`
	Content   string    `db:"content"`
}

/**
 * Set up our DB - make sure it exists and that it has the proper table.
 */
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
}

func addQuote(command *bot.Cmd) (msg string, err error) {
	var insertID int64

	if authorization.Authorize(command.User, command.ChannelData, "author") == false {
		return
	}

	quote := strings.Join(command.Args, " ")
	if quote == "" {
		return
	}

	res, err := db.Exec(
		`INSERT INTO quotes
        (channel, author, timestamp, content)
        VALUES (?, ?, CURRENT_TIMESTAMP, ?)`,
		command.Channel,
		command.User.Nick,
		quote)

	if err != nil {
		err = errors.New("could not insert quote")
		return
	}

	insertID, err = res.LastInsertId()
	if err == nil {
		msg = fmt.Sprintf("Quote inserted with id %d.", insertID)
		log.Println(msg)
	}

	return
}

func queryToQuote(w map[string]interface{}, num uint) (Q Quote) {
	fields := "id, channel, author, timestamp, content"
	query := "SELECT %s FROM quotes WHERE deleted = 0"
	var binds []interface{}
	var usedLike bool

	for where, bind := range w {
		query += fmt.Sprintf(" AND %s", where)
		binds = append(binds, bind)
		if strings.Contains(where, "LIKE") {
			usedLike = true
		}
	}

	var count int
	countRow := db.QueryRow(fmt.Sprintf(query, "COUNT(*)"), binds...)
	err := countRow.Scan(&count)

	if count == 0 || err != nil {
		return
	}

	if count > 1 {
		if num == 0 {
			num = uint(rand.Intn(count)) + 1
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
	var lastArgNum uint

	if len(command.Args) > 0 {
		lastArg := args[len(args)-1]
		lan, err := strconv.Atoi(lastArg)
		if err != nil {
			lastArgNum = uint(0)
			err = nil
		} else {
			lastArgNum = uint(lan)
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

	// todo:
	// - lastquote
	// - delquote (passive, maybe?)
	// - quoteinfo
	// - quotehelp (passive, query only)
}
