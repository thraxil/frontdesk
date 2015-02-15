# Frontdesk

A helpful IRC bot.

## Basic Features

Frontdesk sits in your IRC channel and does some basic helpful things:

* it logs the channel and exposes those logs through a web interface
* it maintains a searchable index of the chat logs and a web interface
  for searching.
* it saves links that are posted to the channel and exposes those
  through a web interface and RSS feed
* if someone in the channel mentions someone that isn't currently
  online, frontdesk takes note and delivers the message to that user
  the next time they come back into the channel.

### Link posting

Where I work, we like to share links with each other in the IRC
channel. Over time, we found it very helpful to have these
archived. Originally, we used the 'phenny' IRC bot and a plugin that
posted the links to a delicious.com account. Delicious seems to be on
the way out (at least, their APIs are changing and
unreliable). Frontdesk replaces that functionality and just keeps
track of links itself. To post a link, just do:

    .url http://example.com/ Title for the Link

Starting it with `.url` is a phenny convention that frontdesk keeps
for the sake of consistency. Frontdesk will see that and send you a
message that your link has been saved. It will then appear on the
recent links page in the web interface and in the RSS feed.

### Off The Record

If you start a line in IRC with `otr:`, front desk will consider it
off the record and not log it.

### Smoketest

Frontdesk exposes a web endpoint with the same output format as
[django-smoketest](https://github.com/ccnmtl/django-smoketest), which
enables us to monitor it along with the rest of our infrastructure (we
used to have a lot of problems with IRC bots dropping offline and no
one noticing). Currently, the only test frontdesk performs is whether
it is currently connected to the IRC server. Frontdesk also notices
when it is disconnected and automatically reconnects (with exponential
backoff), so it should have fewer disconnect issues in general.

## Configuration

Frontdesk is configured 12-factor app style, through environment
variables. Look at the `.env` file for an example. You would typically
run it with a command like:

    $ source .env && ./frontdesk

### FRONTDESK_CHANNEL

The channel for frontdesk to join when it connects to IRC. Include the
`#`

### FRONTDESK_NICK

The nick to use. Try to pick something unique

### FRONTDESK_DB_PATH

Frontdesk uses a boltdb file to store data. This will need to be in a
directory that the user running frontdesk can write to.

### FRONTDESK_BLEVE_PATH

Frontdesk uses [bleve](http://www.blevesearch.com/) for full-text
search indexing. This variable configures the location for bleve to
store its index. Again, needs to be writable by the frontdesk user.

### FRONTDESK_PORT

port for the web interface to listen on

### FRONTDESK_BASE_URL

URL base for links.

### FRONTDESK_HTPASSWD

If this is configured, it will look for an htpasswd file at this
location and use that to set up HTTP Basic auth for the chat
logs. (The assumption is that you still want your links page open to
the public)

## Bugs/Issues

Use github issues to report any issues. Currently, some obvious things
that frontdesk still has some problems with:

* probably ought to be able to set a password and have it identify its
  nick to the IRC server. Not a big deal when it's just logging
  stuff. If it ever grows the ability to manage channel ops, that will
	need to be there though.
* I can't figure out how to get it to reliably track users entering
  and leaving the channel via `JOIN`, `PART`, and `QUIT`. The Go IRC
  library I'm using just doesn't seem to expose those. So frontdesk is
  just polling the channel once per minute. That means that if someone
  leaves or enters, it might be a minute before frontdesk notices. So
  those are two small windows for messages to be missed.
* for similar reasons, frontdesk doesn't do a good job with users
  quickly hopping in and out of the channel. The following scenario
  could happen: user enters channel, frontdesk sees them, sees that it
  has messages to deliver to them, user leaves channel, frontdesk
  sends them the messages that it had (which will fail), frontdesk
  deletes messages from database. Ie, frontdesk needs to do a better
  job of ensuring that the messages were actually delivered to the
  user before deleting them.

## Future Work

* manage ops. add whitelist of nicks and have frontdesk automatically
  grant ops to those nicks when they enter the channel.
* tweet. set up a twitter account for the channel and let anyone (or
  perhaps just a whitelisted set) tweet to it by doing ".tweet this
  goes to twitter"
* email offline messages. let a user tell frontdesk their email
  address and, if someone mentions them while they're not in the
  channel, frontdesk can email them instead of delivering the message
  when they come back online.

## Build/Install

First, [install Go](https://golang.org/doc/install) and make sure you
have your `GOPATH`, etc set up properly (if you can compile a hello
world, you're fine).

Install the dependencies with:

    $ make install_deps

If, for some reason, you don't have `make` available on your system,
you can look in the `Makefile` and just run those `go get` commands
manually.

To compile:

    $ make

Or just

    $ go build .

That will leave a binary names `frontdesk` in the current
directory. Copy it to whereever you want to run the bot and configure
and run it.

If you're doing development, you'll want to test it. For that, there's
a

    $ make run

command that will build it and run it with the configuration specified
in the `.env` file, which points at a testing channel.

You can run the unit tests with

    $ make test

and get a test coverage report with

    $ make coverage

(it will write it out to `coverage.html`)
