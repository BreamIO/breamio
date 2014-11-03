package ircano

import (
	"fmt"
	rs "github.com/maxnordlund/breamio/analysis/regionStats"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"github.com/sorcix/irc"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
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

const nbsp = string('\u00A0')

type Settings struct {
	Username,
	Password,
	Server,
	Channel string
}

type ircBot struct {
	settings  Settings
	stats     []string
	statsLock sync.RWMutex
	conn      *irc.Conn
	closer    chan error
	messages  chan *irc.Message
}

// Implements sort.Interface
type nameSlice []string

var bot *ircBot
var logger *log.Logger

func (n nameSlice) Len() int {
	return len(n)
}

func (n nameSlice) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n nameSlice) Less(i, j int) bool {
	return len(n[i]) < len(n[j])
}

func New() *ircBot {
	return &ircBot{
		stats:    make([]string, 0),
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
						for _, line := range bot.stats {
							bot.send(irc.PRIVMSG, line, bot.settings.Channel)
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
		var regionStatsChan <-chan rs.RegionStatsMap
		var et briee.EventEmitter
		go func() {
			var err error
			time.Sleep(time.Second * 2)
			et, err = l.EmitterLookup(1)
			if err != nil {
				logger.Println(err)
				return
			}
			regionStatsChan = et.Subscribe(RegionStatsEvent, make(rs.RegionStatsMap)).(<-chan rs.RegionStatsMap)
		}()
		settingsChan := l.RootEmitter().Subscribe(SettingsEvent, Settings{}).(<-chan Settings)
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
					names := make(nameSlice, 0, len(regionStats))
					for name, _ := range regionStats {
						names = append(names, name)
					}
					sort.Sort(names)
					max := len(names[len(names)-1])
					stats := []string{
						"EyeStream statistics:",
						sprintf("%-*s | Looks/min | Average look time", max, "Region"),
						strings.Repeat("-", max+32),
					}
					for _, name := range names {
						region := regionStats[name]
						name = strings.ToTitle(name[:1]) + strings.Replace(name[1:], "-", " ", -1)
						stats = append(stats, sprintf(
							"%-*s | %9d | %9.2f seconds",
							max, name,
							region.Looks,
							time.Duration(region.TimeInside).Seconds(),
						))
					}
					bot.statsLock.Lock()
					defer bot.statsLock.Unlock()
					bot.stats = stats
				}(bot, regionStats)
			case <-closer:
				return
			}
		}
	}))
}

func sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
	// return strings.Replace(fmt.Sprintf(format, a...), " ", nbsp, -1)
}
