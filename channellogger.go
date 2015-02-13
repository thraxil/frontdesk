package main

import (
	"encoding/json"
	"fmt"
	"log"

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
	log.Println("logged it...")
}
