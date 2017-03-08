// Package authorization provides user matchting and channel / role authorization. It allows you to dictate who can (and can't) do what.
package authorization

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/go-chat-bot/bot"
)

// Userfile holds our Users and Channels.
type Userfile struct {
	Channels map[string]*Channel `json:"channels"`
	Users    map[string]*User    `json:"users"`
}

// Channel holds a singular Channel with Roles
type Channel struct {
	Roles map[string][]string `json:"roles"`
}

// User holds a singular User with Masks.
type User struct {
	Username string
	Rawmasks []string `json:"masks"`
	Masks    []*regexp.Regexp
}

// Match tries to match a host against known Users.
func (user *User) Match(host string) bool {
	for _, mask := range user.Masks {
		if match := mask.MatchString(host); match {
			return true
		}
	}

	return false
}

func prepareMask(mask string) *regexp.Regexp {
	mask = strings.Replace(mask, ".", "\\.", -1)
	mask = strings.Replace(mask, "*", ".*", -1)
	mask = strings.Replace(mask, "?", ".", -1)
	mask = fmt.Sprintf("(?i)^%s$", mask)
	maskre, err := regexp.Compile(mask)
	if err != nil {
		log.Println("Couldn't turn mask", mask, "into regex:", err)
	}

	return maskre
}

// Fullhost constructs a full (IRC-style) host for a Bot user.
func Fullhost(b *bot.User) string {
	return fmt.Sprintf("%s!%s@%s", b.Nick, b.RealName, b.ID)
}

var (
	userfile Userfile
)

// MatchHost tries to find a hostname, returning a username if found
func MatchHost(host string) (username string, err error) {
	for uname, user := range userfile.Users {
		if user.Match(host) {
			return uname, nil
		}
	}
	return "", errors.New("not found")
}

// Authorize tries to match a channel, role and bot User. It'll return a username if a match is found, or the original nick if a match was made to the special "*" user.
func Authorize(c *bot.ChannelData, r string, b *bot.User) (uname string, err error) {
	if c.IsPrivate {
		return "", errors.New("that's not a channel")
	}

	channel, ok := userfile.Channels[c.Channel]
	if !ok {
		return "", errors.New("channel doesn't exist")
	}

	usernames, ok := channel.Roles[r]
	if !ok {
		return "", fmt.Errorf("role %s doesn't exist in channel", r)
	}

	fullhost := Fullhost(b)
	log.Println("trying full host", fullhost)
	for _, username := range usernames {
		if username == "*" {
			log.Println("anyone can do this")
			return b.Nick, nil
		}

		user, ok := userfile.Users[username]
		if !ok {
			log.Println("user", username, "doesn't exist")
			continue
		}

		if user.Match(fullhost) {
			return username, nil
		}
	}

	return "", errors.New("not found")
}

// Load (re)loads userfile.json into our Userfile struct.
func Load() {
	file, err := ioutil.ReadFile("userfile.json")
	if err != nil {
		log.Fatalf("Could not read file: %#v", err)
	}

	err = json.Unmarshal(file, &userfile)
	if err != nil {
		log.Println("Error while unmarshaling:", err)
	}

	for uname, u := range userfile.Users {
		u.Username = uname
		for _, rm := range u.Rawmasks {
			u.Masks = append(u.Masks, prepareMask(rm))
		}
	}

	log.Println("Loaded userfile.")
	// log.Printf("Userfile: %+v\n", userfile)
}

func reloadUsers(command *bot.Cmd) (msg string, err error) {
	if _, err := Authorize(command.ChannelData, "admin", command.User); err != nil {
		log.Println("Couldn't authorize:", err)
		return "", nil
	}
	Load()
	return
}

func init() {
	Load()
	bot.RegisterCommand(
		"reloadusers",
		"Reload the user file",
		"",
		reloadUsers)
}
