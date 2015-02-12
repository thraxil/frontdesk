package main

import (
	"fmt"
	"log"

	irc "github.com/fluffle/goirc/client"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Channel string
	Nick    string
}

func main() {
	var cfg config
	err := envconfig.Process("frontdesk", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

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

	// this is the handler that gets triggered whenever someone posts
	// in the channel
	c.HandleFunc("PRIVMSG", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println(line.Nick, line.Text())
		if line.Target() == cfg.Channel {
			fmt.Println("to the whole channel")
		} else {
			fmt.Println("to just me")
		}
	})

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

	// Now we can connect
	if err := c.ConnectTo("irc.freenode.net"); err != nil {
		log.Fatal(err.Error())
	}

	// Wait for disconnect
	<-quit
}
