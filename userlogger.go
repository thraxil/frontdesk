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
	running bool
}

func newUserLogger(db *bolt.DB, conn *irc.Conn, channel string) *userLogger {
	ul := &userLogger{db, conn, channel, false}
	go ul.run()
	return ul
}

type nickEntry struct {
	Timestamp time.Time
}

// called when we get a 353 response
func (cl *userLogger) Handle(conn *irc.Conn, line *irc.Line) {
	e := nickEntry{line.Time}

	data, err := json.Marshal(e)
	if err != nil {
		log.Println("error marshalling to json")
		log.Println(err)
		return
	}

	err = cl.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("nicks"))

		for _, n := range strings.Split(line.Text(), " ") {
			err = bucket.Put([]byte(normalizeNick(n)), data)
			if err != nil {
				return err
			}
			log.Println("logged presence of", normalizeNick(n))
		}
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("logged it...")

}

func (cl *userLogger) run() {
	for {
		if cl.running {
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
