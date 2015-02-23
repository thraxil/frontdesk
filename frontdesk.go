package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/abbot/go-http-auth"
	"github.com/blevesearch/bleve"
	"github.com/boltdb/bolt"
	irc "github.com/fluffle/goirc/client"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Channel string
	Nick    string

	DBPath    string `envconfig:"DB_PATH"`
	BlevePath string `envconfig:"BLEVE_PATH"`

	Port         int
	BaseURL      string `envconfig:"BASE_URL"`
	HtpasswdFile string `envconfig:"HTPASSWD"`
	HandleFile   string `envconfig:"HANDLE_FILE"`

	BitlyAccessToken      string `envconfig:"BITLY_ACCESS_TOKEN"`
	TwitterOauthToken     string `envconfig:"TWITTER_OAUTH_TOKEN"`
	TwitterOauthSecret    string `envconfig:"TWITTER_OAUTH_SECRET"`
	TwitterConsumerKey    string `envconfig:"TWITTER_CONSUMER_KEY"`
	TwitterConsumerSecret string `envconfig:"TWITTER_CONSUMER_SECRET"`
}

var backoff = 0
var maxBackoff = 9

func retryConnect(c *irc.Conn) error {
	backoffSecs := time.Duration(
		math.Pow(2, math.Min(float64(backoff), float64(maxBackoff))))
	time.Sleep(backoffSecs * time.Second)
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

	index, err := bleve.Open(cfg.BlevePath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Println("Creating new index")
		indexMapping, err := buildIndexMapping()
		if err != nil {
			log.Fatal(err)
		}
		index, err = bleve.New(cfg.BlevePath, indexMapping)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	} else {
		log.Println("opening existing index")
	}

	c := irc.SimpleClient(cfg.Nick)

	s := newSite(db, index, c, cfg.Channel, cfg.BaseURL, cfg.HtpasswdFile,
		cfg.HandleFile,

		cfg.BitlyAccessToken,

		cfg.TwitterOauthToken, cfg.TwitterOauthSecret,
		cfg.TwitterConsumerKey, cfg.TwitterConsumerSecret,
	)

	// setup IRC handlers
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
	if s.HtpasswdFile != "" {
		log.Println("authentication needed")
		secretProvider := auth.HtpasswdFileProvider(s.HtpasswdFile)
		authenticator := auth.NewBasicAuthenticator("frontdesk", secretProvider)
		http.HandleFunc("/logs/", authenticator.Wrap(makeAuthHandler(logsAuthHandler, s)))
	} else {
		http.HandleFunc("/logs/", makeHandler(logsHandler, s))
	}
	http.HandleFunc("/links/", makeHandler(linksHandler, s))
	http.HandleFunc("/links/feed/", makeHandler(linksFeedHandler, s))
	http.HandleFunc("/search/", makeHandler(searchHandler, s))
	http.HandleFunc("/smoketest/", makeHandler(smoketestHandler, s))
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

func (l lineEntry) Permalink() string {
	return fmt.Sprintf("/logs/%04d/%02d/%02d/#%s", l.Timestamp.Year(),
		l.Timestamp.Month(), l.Timestamp.Day(), l.Key())
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

func makeAuthHandler(fn func(http.ResponseWriter, *auth.AuthenticatedRequest, *site), s *site) auth.AuthenticatedHandlerFunc {
	return func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		log.Printf("%s %s\n", r.Method, r.URL.String())
		fn(w, r, s)
	}
}
