package ircano

import (
	"fmt"
	rs "github.com/maxnordlund/breamio/analysis/regionStats"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/sorcix/irc"
	"log"
	"os"
	"sync"
)

const (
	SettingsEvent    = "ircano:settings"
	RegionStatsEvent = "regionStats:regions"
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
	settings  Settings
	stats     rs.RegionStatsMap
	statsLock sync.RWMutex
	conn      *irc.Conn
	closer    chan error
	messages  chan *irc.Message
}

var bot *ircBot
var logger *log.Logger

func New() *ircBot {
	return &ircBot{
		stats:    make(rs.RegionStatsMap),
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
				switch msg.Trailing {
				case "!Blue circle", "!EyeTracking":
					bot.send(irc.PRIVMSG, "The blue circle on the screen is a "+
						"gazemarker. The gazemarker utilizes EyeStream, a software "+
						"developed by Bream iO, to show my audience where I look on the "+
						"screen. To present statistics on where I look, simply type "+
						"\"!ETStats\".", bot.settings.Channel)
				case "!ETStatsInfo":
					bot.send(irc.PRIVMSG, "Region is the region that the stats is "+
						"collected for. look/min is how many times I look at the region "+
						"per minute. average look time is how long I look at the region "+
						"each time I look there.", bot.settings.Channel)
				case "!ETStats":
					go func(bot *ircBot) {
						bot.statsLock.RLock()
						defer bot.statsLock.RUnlock()
						bot.send(irc.PRIVMSG, "EyeStream statistics:", bot.settings.Channel)
						bot.send(irc.PRIVMSG, "Region\tlooks/min\taverage look time", bot.settings.Channel)
						for name, region := range bot.stats {
							bot.send(irc.PRIVMSG, fmt.Sprintf(
								"%s\t%d\t%d",
								name,
								region.Looks,
								region.TimeInside,
							), bot.settings.Channel)
						}
					}(bot)
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
		logger.Println("Quiting IRC")
		return bot.conn.Close()
	}
	return nil
}

func init() {
	logger = log.New(os.Stderr, "[Ircano] ", log.LstdFlags|log.Lshortfile)
	bot = New()
	bl.Register(bl.NewRunHandler(func(l bl.Logic, closer <-chan struct{}) {
		et, err := l.EmitterLookup(1)
		if err != nil {
			logger.Println(err)
			return
		}
		settingsChan := l.RootEmitter().Subscribe(SettingsEvent, Settings{}).(<-chan Settings)
		regionStatsChan := et.Subscribe(RegionStatsEvent, make(rs.RegionStatsMap)).(<-chan rs.RegionStatsMap)
		defer func() {
			l.RootEmitter().Unsubscribe(SettingsEvent, settingsChan)
			et.Unsubscribe(RegionStatsEvent, regionStatsChan)
			err := bot.Close()
			if err != nil {
				logger.Println(err)
			}
		}()

		for {
			select {
			case settings, ok := <-settingsChan:
				if !ok {
					return
				}
				logger.Println("Settings update")
				bot.Close()
				// Skipping error since we are creating a new connection soon anyway.

				bot.settings = settings
				bot.FillDefaults()
				go bot.ListenAndServe()
			case regionStats, ok := <-regionStatsChan:
				if !ok {
					return
				}
				logger.Println("Region stats update")
				go func(bot *ircBot, regionStats rs.RegionStatsMap) {
					bot.statsLock.Lock()
					defer bot.statsLock.Unlock()
					bot.stats = regionStats
				}(bot, regionStats)
			case <-closer:
				return
			}
		}
	}))
}
