package ircano

import (
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/sorcix/irc"
	"log"
	"os"
)

const (
	SettingsEvent = "ircano:settings"
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

func New() *ircBot {
	return &ircBot{
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
	logger.Println("Connected to", bot.settings.Server)
	bot.conn = conn
	go func() {
		for {
			msg, err := bot.conn.Decode()
			if err != nil {
				logger.Println(err)
				bot.closer <- err
				return
			}
			bot.messages <- msg
		}
	}()

	bot.send(irc.PASS, "", bot.settings.Password)
	bot.send(irc.USER, bot.settings.Username, bot.settings.Username, "0", "*")
	bot.send(irc.NICK, "", bot.settings.Username)
	bot.send(irc.JOIN, "", bot.settings.Channel)

	for {
		select {
		case msg := <-bot.messages:
			switch msg.Command {
			case irc.PING:
				bot.send(irc.PING, msg.Trailing, msg.Params...)
			case irc.PRIVMSG:
				logger.Println("Private message", msg.Trailing)
				if msg.Trailing == "!stats" {
					bot.send(irc.PRIVMSG, "Hello world", bot.settings.Channel)
				}
			}
		case err := <-bot.closer:
			bot.send(irc.QUIT, "")
			return err
		}
	}
}

func (bot *ircBot) send(command, trailing string, params ...string) error {
	logger.Println("Sending", command, params, trailing)
	err := bot.conn.Encode(&irc.Message{
		Command:  command,
		Params:   params,
		Trailing: trailing,
	})
	if err != nil {
		logger.Println(err)
	}
	return err
}

func (bot *ircBot) Close() error {
	if bot.conn != nil {
		bot.closer <- nil
		return bot.conn.Close()
	}
	return nil
}

func init() {
	logger = log.New(os.Stderr, "[Ircano] ", log.LstdFlags|log.Lshortfile)
	bot = New()
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
				logger.Println("Settings update")
				bot.Close()
				// Skipping error since we are creating a new connection soon anyway.

				bot.settings = s
				bot.FillDefaults()
				go bot.ListenAndServe()
			case <-closer:
				logger.Println("Quiting IRC")
				return
			}
		}
	}))
}
