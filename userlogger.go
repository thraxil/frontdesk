package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	irc "github.com/fluffle/goirc/client"
)

type userLogger struct {
	db      *bolt.DB
	conn    *irc.Conn
	channel string
	site    *site
	running bool
}

func newUserLogger(db *bolt.DB, conn *irc.Conn, channel string, site *site) *userLogger {
	ul := &userLogger{db, conn, channel, site, false}
	go ul.run()
	return ul
}

type nickEntry struct {
	Timestamp time.Time
}

// called when we get a 353 response
func (cl *userLogger) Handle(conn *irc.Conn, line *irc.Line) {
	previous := cl.site.onlineNicks()
	e := nickEntry{line.Time}

	data, err := json.Marshal(e)
	if err != nil {
		log.Println("error marshalling to json")
		log.Println(err)
		return
	}
	nicks := strings.Split(line.Text(), " ")
	for _, n := range nicks {
		_, ok := previous[normalizeNick(n)]
		if !ok {
			log.Println(n, "has entered the channel")
			log.Println("this is where we let them know if they had any messages")
		}
	}
	err = cl.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("nicks"))

		for _, n := range nicks {
			err = bucket.Put([]byte(normalizeNick(n)), data)
			if err != nil {
				return err
			}
		}

		online := tx.Bucket([]byte("online"))
		err = online.Put([]byte("now"), []byte(strings.Join(nicks, " ")))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (cl *userLogger) run() {
	for {
		if cl.running && cl.conn != nil {
			// request list of current nicks
			cl.conn.Raw("NAMES" + " " + cl.channel)
		}
		time.Sleep(60 * time.Second)
	}
}

func (cl *userLogger) start() {
	cl.running = true
}

func (cl *userLogger) stop() {
	cl.running = false
}
