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

func (s *Settings) FillDefaults() {
	if s.Username == "" {
		s.Username = Username
		if s.Password == "" {
			s.Password = Password
		}
	}
	if s.Server == "" {
		s.Server = Server
	}
}

type IRCBot struct {
	settings Settings
	*irc.Conn
}

var bot *IRCBot
var logger *log.Logger

func New(s Settings) *IRCBot {
	return &IRCBot{s, nil}
}

func (bot IRCBot) ListenAndServe() {

}

func (bot IRCBot) Close() error {
	return nil
}

func init() {
	logger = log.New(os.Stderr, "[Ircano] ", log.LstdFlags|log.Lshortfile)
	bot = &IRCBot{}
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
				s.FillDefaults()
				bot.Close()
				//Skipping error because we do not care since we are creating a new connection soon anyway.

				bot = New(s)
				go bot.ListenAndServe()
			case <-closer:
				return
			}
		}
	}))
}
