package authorization

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"

	"github.com/go-chat-bot/bot"
)

// Userfile holds our Users and Channels
type Userfile struct {
	Users    map[string]User    `json:"users"`
	Channels map[string]Channel `json:"channels"`
}

// User holds a singular User with Masks
type User struct {
	Masks []string `json:"masks"`
}

// Match tries to match a host against known Users.
func (user *User) Match(host string) bool {
	for _, mask := range user.Masks {
		match, err := regexp.MatchString(fmt.Sprintf("(?i)^%s$", mask), host)
		if err != nil {
			log.Println(err)
			continue
		}
		if match {
			return true
		}
	}
	return false
}

// Channel holds a singular Channel with Roles
type Channel struct {
	Roles map[string][]string `json:"roles"`
}

var (
	userfile Userfile
)

// HostFromUser constructs a full user host mask, inpsired by IRC, ready for matching against a regex.
func HostFromUser(user *bot.User) string {
	return fmt.Sprintf("%s!%s@%s", user.Nick, user.RealName, user.ID)
}

// MatchHost tries to find a hostname, returning a username if found
func MatchHost(host string) (username string, err error) {
	for uname, user := range userfile.Users {
		if user.Match(host) {
			return uname, nil
		}
	}
	return "", errors.New("not found")
}

// Authorize tries to match a channel, role and bot User.
func Authorize(c *bot.ChannelData, role string, usr *bot.User) (uname string, err error) {
	if c.IsPrivate {
		return "", errors.New("that's not a channel")
	}
	channame, ok := userfile.Channels[c.Channel]
	if !ok {
		return "", errors.New("channel doesn't exist")
	}
	chanrole, ok := channame.Roles[role]
	if !ok {
		return "", errors.New("role doesn't exist in channel")
	}
	host := HostFromUser(usr)
	uname, err = MatchHost(host)
	if err != nil {
		return "", errors.New("couldn't match host")
	}
	for _, username := range chanrole {
		user, ok := userfile.Users[username]
		if !ok {
			continue
		}
		if user.Match(host) {
			return uname, nil
		}
	}
	return "", errors.New("not found")
}

// ReloadUserfile reloads userfile.json into our Userfile struct.
func ReloadUserfile() Userfile {
	file, err := ioutil.ReadFile("userfile.json")
	if err != nil {
		log.Fatalf("Couldn't read file: %#v", err)
	}

	json.Unmarshal(file, &userfile)
	log.Println("Loaded userfile")

	return userfile
}

func reloadUsers(command *bot.Cmd) (msg string, err error) {
	if _, err := Authorize(command.ChannelData, "admin", command.User); err != nil {
		return "", nil
	}

	ReloadUserfile()

	return
}

func init() {
	ReloadUserfile()

	bot.RegisterCommand(
		"reloadusers",
		"Reload the user file",
		"",
		reloadUsers)
}
