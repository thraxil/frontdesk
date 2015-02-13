package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/boltdb/bolt"
	irc "github.com/fluffle/goirc/client"
)

type site struct {
	channelLogger *channelLogger
	userLogger    *userLogger
	db            *bolt.DB
}

func newSite(db *bolt.DB, conn *irc.Conn, channel string) *site {
	s := &site{db: db}
	cl := newChannelLogger(db, channel, s)
	ul := newUserLogger(db, conn, channel, s)
	s.channelLogger = cl
	s.userLogger = ul
	s.ensureBuckets()
	return s
}

func (s *site) ensureBuckets() {
	err := s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("lines"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("online"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("nicks"))
		return err
	})
	if err != nil {
		log.Fatal("couldn't ensure existence of basic buckets")
	}
	return
}

func (s site) years() []string {
	years := []string{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		b.ForEach(func(k, v []byte) error {
			years = append(years, string(k))
			return nil
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return years
}

func (s site) linesForDay(year, month, day string) []lineEntry {
	entries := []lineEntry{}
	err := s.db.View(func(tx *bolt.Tx) error {
		lb := tx.Bucket([]byte("lines"))
		yb := lb.Bucket([]byte(year))
		mb := yb.Bucket([]byte(month))
		db := mb.Bucket([]byte(day))
		db.ForEach(func(k, v []byte) error {
			var e lineEntry
			err := json.Unmarshal(v, &e)
			if err != nil {
				return err
			}
			entries = append(entries, e)
			return nil
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return entries
}

func (s site) daysForMonth(year, month string) []string {
	entries := []string{}
	err := s.db.View(func(tx *bolt.Tx) error {
		lb := tx.Bucket([]byte("lines"))
		yb := lb.Bucket([]byte(year))
		mb := yb.Bucket([]byte(month))
		mb.ForEach(func(k, v []byte) error {
			entries = append(entries, string(k))
			return nil
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return entries
}

func (s site) monthsForYear(year string) []string {
	entries := []string{}
	err := s.db.View(func(tx *bolt.Tx) error {
		lb := tx.Bucket([]byte("lines"))
		yb := lb.Bucket([]byte(year))
		yb.ForEach(func(k, v []byte) error {
			entries = append(entries, string(k))
			return nil
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return entries
}

// all the nicks we've ever seen.
func (s site) allKnownNicks() []string {
	nicks := []string{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nicks"))
		b.ForEach(func(k, v []byte) error {
			nicks = append(nicks, string(k))
			return nil
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return nicks
}

func (s site) onlineNicks() map[string]bool {
	nicks := map[string]bool{}

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("online"))
		v := b.Get([]byte("now"))
		if v == nil {
			return nil
		}
		for _, n := range strings.Split(string(v), " ") {
			nicks[normalizeNick(n)] = true
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return nicks
}
