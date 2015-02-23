package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/boltdb/bolt"
	irc "github.com/fluffle/goirc/client"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/thraxil/bitly"
	"github.com/xiam/twitter"
)

type site struct {
	channelLogger *channelLogger
	userLogger    *userLogger
	db            *bolt.DB
	index         bleve.Index
	BaseURL       string
	HtpasswdFile  string

	BitlyAccessToken      string
	TwitterOauthToken     string
	TwitterOauthSecret    string
	TwitterConsumerKey    string
	TwitterConsumerSecret string
}

func newSite(db *bolt.DB, index bleve.Index, conn *irc.Conn, channel, baseURL, htpasswdFile,
	bitlyAccessToken, twitterOauthToken, twitterOauthSecret, twitterConsumerKey,
	twitterConsumerSecret string) *site {
	s := &site{
		db: db, index: index, BaseURL: baseURL, HtpasswdFile: htpasswdFile,
		BitlyAccessToken:      bitlyAccessToken,
		TwitterOauthToken:     twitterOauthToken,
		TwitterOauthSecret:    twitterOauthSecret,
		TwitterConsumerKey:    twitterConsumerKey,
		TwitterConsumerSecret: twitterConsumerSecret,
	}
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
		_, err = tx.CreateBucketIfNotExists([]byte("links"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("mentions"))
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

func (s site) getLines(keys []string) []lineEntry {
	entries := []lineEntry{}
	err := s.db.View(func(tx *bolt.Tx) error {
		lb := tx.Bucket([]byte("lines"))
		for _, k := range keys {
			t, err := time.Parse(time.RFC3339Nano, k)
			if err != nil {
				continue
			}
			yb := lb.Bucket([]byte(fmt.Sprintf("%04d", t.Year())))
			mb := yb.Bucket([]byte(fmt.Sprintf("%02d", t.Month())))
			db := mb.Bucket([]byte(fmt.Sprintf("%02d", t.Day())))
			v := db.Get([]byte(k))
			if v == nil {
				continue
			}
			var e lineEntry
			err = json.Unmarshal(v, &e)
			if err != nil {
				continue
			}
			entries = append(entries, e)
		}
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

func (s site) offlineNicks() []string {
	allNicks := s.allKnownNicks()
	onlineNicks := s.onlineNicks()
	offlineNicks := []string{}
	for _, n := range allNicks {
		_, ok := onlineNicks[n]
		if !ok {
			offlineNicks = append(offlineNicks, n)
		}
	}
	return offlineNicks
}

type linkEntry struct {
	Nick      string
	URL       string
	Title     string
	Year      int
	Month     int
	Day       int
	Key       string
	Timestamp time.Time
}

func (e linkEntry) FormattedTimestamp() string {
	return e.Timestamp.Format("Mon Jan 2 15:04:05")
}

func (e linkEntry) DiscussionLink() string {
	return fmt.Sprintf("/logs/%04d/%02d/%02d/#%s", e.Year, e.Month, e.Day, e.Key)
}

func (s *site) saveLink(line *irc.Line, url, title string) {
	year, month, day := line.Time.Date()
	key := line.Time.Format(time.RFC3339Nano)
	le := linkEntry{normalizeNick(line.Nick), url, title,
		year, int(month), day, key, line.Time}
	data, err := json.Marshal(le)
	if err != nil {
		log.Println("error marshalling to json")
		log.Println(err)
		return
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("links"))
		return bucket.Put([]byte(key), data)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (s site) shortenLink(url string) string {
	if s.BitlyAccessToken == "" {
		// no bitly access key. can't shorten
		return url
	}
	c := bitly.NewConnection(s.BitlyAccessToken)
	bitlyLink, err := c.Shorten(url)
	if err != nil {
		log.Println(err)
		// TODO: inform the user that we couldn't shorten their link
		return url
	}
	return strings.Trim(bitlyLink, "\n\r ")
}

func twitterHandleFromNick(nick string) string {
	// TODO: not really implemented yet.
	// need to get the list from a config file.
	// for now, just stick an '@` on there...
	return "@" + nick
}

func (s *site) tweetLink(nick, url, title string) {
	if s.TwitterOauthToken == "" {
		// twitter isn't configured, so we can't tweet
		return
	}
	// shorten the link
	url = s.shortenLink(url)
	log.Println("shortened:", url)

	client := twitter.New(&oauth.Credentials{
		s.TwitterConsumerKey,
		s.TwitterConsumerSecret,
	})
	client.SetAuth(&oauth.Credentials{
		s.TwitterOauthToken,
		s.TwitterOauthSecret,
	})
	_, err := client.VerifyCredentials(nil)
	if err != nil {
		log.Println("twitter credentials are bad")
		log.Println(err)
		// TODO: let user know that credentials are bad
		return
	}
	_, err = client.Update(fmt.Sprintf("%s: %s via %s", url, title,
		twitterHandleFromNick(normalizeNick(nick))), nil)
	if err != nil {
		log.Println("failed to tweet")
		log.Println(err)
		return
	}
	log.Println("successfully tweeted")
}

func (s site) recentLinks() []linkEntry {
	links := []linkEntry{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("links"))
		c := b.Cursor()
		cnt := 0
		for k, v := c.Last(); k != nil && cnt < 100; k, v = c.Prev() {
			var le linkEntry
			err := json.Unmarshal(v, &le)
			if err != nil {
				return err
			}
			links = append(links, le)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return links
}

func (s *site) deliverMessages(nick string, conn *irc.Conn) {
	messages := []mention{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("mentions"))
		v := b.Get([]byte(normalizeNick(nick)))
		if v == nil {
			return nil
		}
		var ms mentions
		err := json.Unmarshal(v, &ms)
		if err != nil {
			return err
		}
		messages = ms.Mentions
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(messages) == 0 {
		return
	}
	// notify them
	conn.Privmsg(nick, fmt.Sprintf("messages while you were out: %d", len(messages)))
	for _, m := range messages {
		conn.Privmsg(nick, fmt.Sprintf("from %s: %s", m.Nick, m.Text))
		conn.Privmsg(nick, "<"+s.BaseURL+m.Permalink()+">")
	}
	// then clear them out
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("mentions"))
		return b.Delete([]byte(normalizeNick(nick)))
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (s *site) indexLine(line *irc.Line) {
	le := lineEntry{normalizeNick(line.Nick), line.Text(), line.Time}
	err := s.index.Index(le.Key(), le)

	log.Println(s.index.DocCount())
	if err != nil {
		log.Fatal(err)
	}
}
