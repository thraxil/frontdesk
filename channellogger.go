package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	irc "github.com/fluffle/goirc/client"
)

type channelLogger struct {
	db      *bolt.DB
	channel string
	site    *site
}

func newChannelLogger(db *bolt.DB, channel string, site *site) *channelLogger {
	return &channelLogger{db, channel, site}
}

func (cl *channelLogger) Handle(conn *irc.Conn, line *irc.Line) {
	if strings.HasPrefix(line.Text(), "otr:") {
		// this is off the record
		return
	}
	if line.Target() == cl.channel {
		cl.logLine(line)
		go cl.saveUrls(conn, line)
		go cl.saveMentions(conn, line)
	} else {
		// process it for commands
	}
}

func (cl *channelLogger) saveUrls(conn *irc.Conn, line *irc.Line) {
	if !strings.HasPrefix(line.Text(), ".url") {
		return
	}
	parts := strings.Split(line.Text(), " ")
	if len(parts) == 1 {
		conn.Privmsg(line.Nick, "syntax: .url http://example.com/ title for link")
		return
	}
	url := parts[1]
	if !strings.HasPrefix(url, "http") {
		// doesn't look like a URL
		conn.Privmsg(line.Nick, fmt.Sprintf("%s doesn't look like a URL", url))
		return
	}
	if len(parts) == 2 {
		conn.Privmsg(line.Nick, "syntax: .url http://example.com/ title for link")
		return
	}
	title := strings.Join(parts[2:], " ")
	cl.site.saveLink(line, url, title)
	conn.Privmsg(line.Nick, "saved your link")
}

func (cl *channelLogger) saveMentions(conn *irc.Conn, line *irc.Line) {
	nicksToCheck := cl.site.offlineNicks()
	for _, n := range nicksToCheck {
		if (strings.Contains(line.Text(), n + ": ") ||
			strings.Contains(line.Text(), n + " ")
		) {
			// offline user was mentioned
			cl.saveMention(n, line, conn)
		}
	}
}

type mention struct {
	Nick      string    `json:"nick"`
	Year      int       `json:"year"`
	Month     int       `json:"month"`
	Day       int       `json:"day"`
	Key       string    `json:"key"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

func (m mention) Permalink() string {
	return fmt.Sprintf("/logs/%04d/%02d/%02d/#%s", m.Year, m.Month, m.Day, m.Key)
}

type mentions struct {
	Mentions []mention `json:"mentions"`
}

func (cl *channelLogger) saveMention(nick string, line *irc.Line, conn *irc.Conn) {
	year, month, day := line.Time.Date()
	key := line.Time.Format(time.RFC3339Nano)
	m := mention{normalizeNick(line.Nick), year,
		int(month), day, key, line.Text(), line.Time}

	var ms mentions
	err := cl.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("mentions"))
		v := bucket.Get([]byte(nick))
		if v == nil {
			// create
			ms.Mentions = []mention{m}
		} else {
			// update
			err := json.Unmarshal(v, &ms)
			if err != nil {
				ms.Mentions = []mention{m}
			} else {
				ms.Mentions = append(ms.Mentions, m)
			}
		}
		data, err := json.Marshal(ms)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(nick), data)
	})
	if err != nil {
		log.Fatal(err)
	}
	conn.Privmsg(line.Nick, fmt.Sprintf("%s is not in the channel right now, but I'll deliver your message when they return", nick))
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
		lines := tx.Bucket([]byte("lines"))
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
	cl.site.indexLine(line)
}
