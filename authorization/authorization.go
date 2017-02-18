package authorization

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"

	"github.com/go-chat-bot/bot"
)

var (
	userfile Userfile
)

// Userfile reflects the entire JSON file.
type Userfile struct {
	Users    []User    `json:"users"`
	Channels []Channel `json:"channels"`
}

// User holds a singular user.
type User struct {
	Username string   `json:"username"`
	Masks    []string `json:"masks"`
}

// Channel holds channel name and various user types and their Users.
type Channel struct {
	Channel string `json:"channel"`
	Roles   []Role
}

// Role is a type of action a User can do on a Channel.
type Role struct {
	Role  string
	Users []string
}

// Authorize a bot User against a role in a channel.
func Authorize(user *bot.User, channel *bot.ChannelData, role string) bool {
	username, err := FindUsername(FullHostmask(user))
	if err != nil {
		return false
	}

	for _, Channel := range userfile.Channels {
		if channel.IsPrivate || channel.Channel != Channel.Channel {
			continue
		}

		var usersInRole []string
		for _, crole := range Channel.Roles {
			if crole.Role != role {
				continue
			} else {
				usersInRole = crole.Users
			}
		}

		for _, uname := range usersInRole {
			if uname == "*" || uname == username {
				return true
			}
		}

		return false
	}

	return false
}

// FullHostmask constructs a full user host mask, inpsired by IRC, ready for matching against a regex.
func FullHostmask(user *bot.User) []byte {
	return []byte(fmt.Sprintf("%s!%s@%s", user.Nick, user.RealName, user.ID))
}

// FindUsername finds a username for a given full host.
func FindUsername(host []byte) (username string, err error) {
	for _, user := range userfile.Users {
		for _, mask := range user.Masks {
			match, err := regexp.Match(fmt.Sprintf("^%s$", mask), host)
			if err != nil {
				return "", err
			}
			if match {
				username = user.Username
				return username, nil
			}

		}
	}

	return "*", nil
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
	if Authorize(command.User, command.ChannelData, "admin") == false {
		return
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
