package main

import (
	"encoding/json"
	"fmt"
	"log"
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

	c.HandleFunc("connected", func(conn *irc.Conn, line *irc.Line) {
		conn.Join(cfg.Channel)
		fmt.Println("connected to the channel", cfg.Channel, "as", cfg.Nick)
		conn.Raw("NAMES" + " " + cfg.Channel)
	})

	quit := make(chan bool)
	c.HandleFunc("disconnected", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("disconnecting")
		quit <- true
	})

	cl := newChannelLogger(db, cfg.Channel)
	s := newSite(cl, db)

	// this is the handler that gets triggered whenever someone posts
	// in the channel
	c.Handle("PRIVMSG", cl)

	c.HandleFunc("NOTICE", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("NOTICE", line.Nick, line.Text())
	})

	c.HandleFunc("353", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("353", line.Nick, line.Text())
		fmt.Println(line.Raw)
	})

	c.HandleFunc("301", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("301 RPL_AWAY", line.Nick, line.Text())
		fmt.Println(line.Raw)
	})

	c.HandleFunc("305", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("305 RPL_UNAWAY", line.Nick, line.Text())
		fmt.Println(line.Raw)
	})

	c.HandleFunc("306", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("306 RPL_NOAWAY", line.Nick, line.Text())
		fmt.Println(line.Raw)
	})

	c.HandleFunc("ACTION", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("ACTION", line.Nick, line.Text())
	})

	c.HandleFunc("QUIT", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("QUIT", line.Nick, line.Text())
	})

	c.HandleFunc("JOIN", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("JOIN", line.Nick, line.Text())
	})

	c.HandleFunc("PART", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("PART", line.Nick, line.Text())
	})

	c.HandleFunc("AWAY", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("AWAY", line.Nick, line.Text())
	})

	c.HandleFunc("MODE", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("MODE", line.Nick, line.Text())
	})

	http.HandleFunc("/", makeHandler(indexHandler, s))
	http.HandleFunc("/logs/", makeHandler(logsHandler, s))
	http.HandleFunc("/favicon.ico", faviconHandler)

	// Now we can connect
	if err := c.ConnectTo("irc.freenode.net"); err != nil {
		log.Fatal(err.Error())
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))

	// Wait for disconnect
	<-quit
}

type channelLogger struct {
	db      *bolt.DB
	channel string
}

func newChannelLogger(db *bolt.DB, channel string) *channelLogger {
	return &channelLogger{db, channel}
}

func (cl *channelLogger) Handle(conn *irc.Conn, line *irc.Line) {
	fmt.Println(line.Nick, line.Text())
	fmt.Println(line.Time)
	if line.Target() == cl.channel {
		fmt.Println("to the whole channel")
		cl.logLine(line)
	} else {
		fmt.Println("to just me")
		// process it for commands
	}
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

func (cl *channelLogger) logLine(line *irc.Line) {
	year, month, day := line.Time.Date()
	le := lineEntry{normalizeNick(line.Nick), line.Text(), line.Time}
	data, err := json.Marshal(le)
	if err != nil {
		log.Println("error marshalling to json")
		log.Println(err)
		return
	}

	err = cl.db.Update(func(tx *bolt.Tx) error {
		lines, err := tx.CreateBucketIfNotExists([]byte("lines"))
		if err != nil {
			return err
		}
		ybucket, err := lines.CreateBucketIfNotExists([]byte(fmt.Sprintf("%04d", year)))
		if err != nil {
			return err
		}
		mbucket, err := ybucket.CreateBucketIfNotExists([]byte(fmt.Sprintf("%02d", month)))
		if err != nil {
			return err
		}
		dbucket, err := mbucket.CreateBucketIfNotExists([]byte(fmt.Sprintf("%02d", day)))
		if err != nil {
			return err
		}
		err = dbucket.Put([]byte(le.Key()), data)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("logged it...")
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
