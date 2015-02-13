package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

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
		go cl.saveMentions(line)
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

func (cl *channelLogger) saveMentions(line *irc.Line) {

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
}
