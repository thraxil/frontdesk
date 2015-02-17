package main

import (
	"testing"
	"time"
)

func Test_linkEntryFormattedTimestamp(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339Nano, "2015-02-15T12:04:36.439011141-05:00")
	le := linkEntry{
		"nick",
		"http://foo.com/",
		"a title",
		2015, 02, 17,
		"a-key",
		ts,
	}
	if le.FormattedTimestamp() != "Sun Feb 15 12:04:36" {
		t.Error(le.FormattedTimestamp())
	}
}

func Test_linkEntryDiscussionLink(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339Nano, "2015-02-15T12:04:36.439011141-05:00")
	le := linkEntry{
		"nick",
		"http://foo.com/",
		"a title",
		2015, 02, 17,
		"a-key",
		ts,
	}
	if le.DiscussionLink() != "/logs/2015/02/17/#a-key" {
		t.Error(le.DiscussionLink())
	}
}
