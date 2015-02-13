package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	irc "github.com/fluffle/goirc/client"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Channel string
	Nick    string

	DBPath string `envconfig:"DB_PATH"`

	Port int
}

var backoff = 0
var max_backoff = 9

func retryConnect(c *irc.Conn) error {
	backoff_secs := time.Duration(
		math.Pow(2, math.Min(float64(backoff), float64(max_backoff))))
	time.Sleep(backoff_secs * time.Second)
	// connect to irc
	if err := c.ConnectTo("irc.freenode.net"); err != nil {
		log.Println("connection attempt", backoff, err.Error())
		backoff++
		return err
	}
	backoff = 0
	return nil
}

func connect(c *irc.Conn) {
	for {
		err := retryConnect(c)
		if err == nil {
			return
		}
	}
}

func main() {
	var cfg config
	err := envconfig.Process("frontdesk", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := bolt.Open(cfg.DBPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	c := irc.SimpleClient(cfg.Nick)

	s := newSite(db, c, cfg.Channel)

	c.HandleFunc("connected", func(conn *irc.Conn, line *irc.Line) {
		conn.Join(cfg.Channel)
		log.Println("connected to the channel", cfg.Channel, "as", cfg.Nick)
		s.userLogger.start()
	})

	c.HandleFunc("disconnected", func(conn *irc.Conn, line *irc.Line) {
		log.Println("disconnecting")
		s.userLogger.stop()
		connect(c)
	})

	// this is the handler that gets triggered whenever someone posts
	// in the channel
	c.Handle("PRIVMSG", s.channelLogger)
	// 353 is the response to a NAMES query
	c.Handle("353", s.userLogger)

	// a bunch more IRC commands that we just want to print
	// to the console if we see them
	cmds := []string{"NOTICE", "301", "305", "306", "ACTION",
		"QUIT", "JOIN", "PART", "AWAY", "MODE"}

	for _, cmd := range cmds {
		c.HandleFunc(cmd, func(conn *irc.Conn, line *irc.Line) {
			log.Println(cmd, line.Nick, line.Text())
		})
	}

	// set up our web handlers
	http.HandleFunc("/", makeHandler(indexHandler, s))
	http.HandleFunc("/logs/", makeHandler(logsHandler, s))
	http.HandleFunc("/favicon.ico", faviconHandler)

	// connect to irc
	connect(c)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))
}

type lineEntry struct {
	Nick      string
	Text      string
	Timestamp time.Time
}

func (l lineEntry) Key() string {
	return l.Timestamp.Format(time.RFC3339Nano)
}

func (l lineEntry) NiceTime() string {
	return l.Timestamp.Format("15:04:05")
}

// IRC likes to rename 'foo' to 'foo_', etc.
// we need to consider them the same
func normalizeNick(n string) string {
	return strings.TrimRight(n, "_")
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *site), s *site) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL.String())
		fn(w, r, s)
	}
}
