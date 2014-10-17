package main

import (
	"fmt"
	ircevent "github.com/thoj/go-ircevent"
	"time"
)

/**
 * TODO:
 * - Add config loader, `config.json`, for username/password
 * - Add callback for commands
 * - Add TCP connection to local breamio server
 * - Add !stats command using the above
 */

func main() {
	fmt.Println("IRCano", Version, GitSHA)
	irc := ircevent.IRC("breamio", "breamio")
	irc.Password = "oauth:<insert this>"
	irc.AddCallback("001", func(event *ircevent.Event) {
		fmt.Println(
			event.Code,
			event.Raw,
			event.Nick,
			event.Host,
			event.Source,
			event.User,
			event.Arguments,
		)
	})
	irc.Connect("irc.twitch.tv:6667")
	time.Sleep(time.Second * 1)
	irc.Join("#maxnordlund")
	irc.Privmsg("#maxnordlund", "Hello World")
	irc.Privmsg("@maxnordlund", "Hello Private World")
	time.Sleep(time.Second * 1)
	irc.Quit()
}
