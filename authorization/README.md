# Authorization

Expects a file `userfile.json` to be present in your bot's working directory. See `userfile_dist.json` for a skeleton file. The general idea is that you define users, which have host masks, and channels, which have roles, which map to usernames.

You can define any role you want, and hook up some usernames to it. In a plugin you can then authorize an action like so:

```go
if authorization.Authorize(command.User, command.Channel, "author") == true {
	// Yep, this user is allowed to do that.
}
```