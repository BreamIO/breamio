package ircano

import (
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/sorcix/irc"
	"log"
	"os"
)

const (
	SettingsEvent = "irc:settings"
)

const (
	Username = "breamio"
	Password = "oauth:rizyskb9a694burwouzh71tsmqxr6lx"
	Server   = "irc.twitch.tv:6667"
)

type Settings struct {
	Username,
	Password,
	Server,
	Channel string
}

type ircBot struct {
	settings Settings
	conn     *irc.Conn
	closer   chan error
	messages chan *irc.Message
}

var bot *ircBot
var logger *log.Logger

func New(s Settings) *ircBot {
	return &ircBot{
		settings: s,
		closer:   make(chan error),
		messages: make(chan *irc.Message),
	}
}

func (bot *ircBot) FillDefaults() {
	if bot.settings.Username == "" {
		bot.settings.Username = Username
		if bot.settings.Password == "" {
			bot.settings.Password = Password
		}
	}
	if bot.settings.Server == "" {
		bot.settings.Server = Server
	}
}

func (bot *ircBot) ListenAndServe() error {
	conn, err := irc.Dial(bot.settings.Server)
	if err != nil {
		return err
	}
	bot.conn = conn
	go func() {
		msg, err := bot.conn.Decode()
		if err != nil {
			bot.closer <- err
			return
		}
		bot.messages <- msg
	}()
	bot.conn.Encode(&irc.Message{
		Prefix: &irc.Prefix{
			Name: bot.settings.Username,
			User: bot.settings.Username,
		},
		Command: irc.PASS,
		Params:  []string{bot.settings.Password},
	})
	for {
		select {
		case msg := <-bot.messages:
			switch msg.Command {
			case irc.PRIVMSG:
				if msg.Trailing == "!stats" {
					bot.conn.Encode(&irc.Message{
						Prefix: &irc.Prefix{
							Name: bot.settings.Username,
							User: bot.settings.Username,
						},
						Command:  irc.PRIVMSG,
						Params:   []string{string(irc.Channel) + bot.settings.Channel},
						Trailing: "Hello World",
					})
				}
			}
		case err := <-bot.closer:
			return err
		}
	}
}

func (bot *ircBot) Close() error {
	bot.closer <- nil
	return bot.conn.Close()
}

func init() {
	logger = log.New(os.Stderr, "[Ircano] ", log.LstdFlags|log.Lshortfile)
	bot = &ircBot{}
	bl.Register(bl.NewRunHandler(func(l bl.Logic, closer <-chan struct{}) {
		settings := l.RootEmitter().Subscribe(SettingsEvent, Settings{}).(<-chan Settings)
		defer func() {
			l.RootEmitter().Unsubscribe(SettingsEvent, settings)
			err := bot.Close()
			if err != nil {
				logger.Println(err)
			}
		}()

		for {
			select {
			case s, ok := <-settings:
				if !ok {
					return
				}
				bot.Close()
				// Skipping error since we are creating a new connection soon anyway.

				bot.settings = s
				bot.FillDefaults()
				go bot.ListenAndServe()
			case <-closer:
				return
			}
		}
	}))
}
