package ircano

import (
	rs "github.com/maxnordlund/breamio/analysis/regionStats"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"

	"github.com/sorcix/irc"

	"fmt"
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

type Factory struct{}

func (Factory) String() string {
	return "IRC"
}

func (f Factory) New(c bl.Constructor) bl.Module {
	bot := &ircBot{
		SimpleModule: bl.NewSimpleModule(f.String(), c),
		stats:        make([]string, 0),
		closer:       make(chan error),
		messages:     make(chan *irc.Message),
	}

	go bot.ListenAndServe()
	return bot
}

type ircBot struct {
	bl.SimpleModule
	settings  Settings
	stats     []string
	statsLock sync.RWMutex
	conn      *irc.Conn
	closer    chan error
	messages  chan *irc.Message

	MethodOnRegionStats bl.EventMethod `event:"RegionStats:regions"`
}

// Implements sort.Interface
type nameSlice []string

func (n nameSlice) Len() int {
	return len(n)
}

func (n nameSlice) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n nameSlice) Less(i, j int) bool {
	return len(n[i]) < len(n[j])
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

func (bot *ircBot) OnRegionStats() {
	logger.Println("Region stats update")

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
}

func (bot *ircBot) ListenAndServe() error {
	conn, err := irc.Dial(bot.settings.Server)
	if err != nil {
		return err
	}
	bot.Logger().Println("Connected to", bot.settings.Server)
	bot.conn = conn
	go func() {
		for {
			msg, err := bot.conn.Decode()
			if err != nil {
				bot.Logger().Println(err)
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

	return bot.serve()
}

func (bot *ircBot) serve() error {
	for {
		select {
		case msg := <-bot.messages:
			switch msg.Command {
			case irc.PING:
				bot.send(irc.PING, msg.Trailing, msg.Params...)
			case irc.PRIVMSG:
				bot.Logger().Println("Private message", msg.Trailing)
				switch msg.Trailing {
				case "!Blue circle", "!EyeTracking":
					bot.send(irc.PRIVMSG, infoMessage, bot.settings.Channel)
				case "!ETStatsInfo":
					bot.send(irc.PRIVMSG, explanation, bot.settings.Channel)
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
	bot.Logger().Println("Sending", command, params, trailing)
	err := bot.conn.Encode(&irc.Message{
		Command:  command,
		Params:   params,
		Trailing: trailing,
	})
	if err != nil {
		bot.Logger().Println(err)
	}
	return err
}

func (bot *ircBot) Close() error {
	if bot.conn != nil {
		close(bot.closer)
		bot.Logger().Println("Quiting IRC")
		return bot.conn.Close()
	}
	return nil
}

func init() {
	bl.Register(Factory{})
}

func sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
	// return strings.Replace(fmt.Sprintf(format, a...), " ", nbsp, -1)
}
