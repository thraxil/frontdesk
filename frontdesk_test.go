package main

import (
	"testing"
	"time"
)

func Test_lineEntryKey(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339Nano, "2015-02-15T12:04:36.439011141-05:00")
	le := lineEntry{Nick: "foo", Text: "blah", Timestamp: ts}
	k := le.Key()
	if k != "2015-02-15T12:04:36.439011141-05:00" {
		t.Error("wrong key")
	}
}

func Test_lineEntryNiceTime(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339Nano, "2015-02-15T12:04:36.439011141-05:00")
	le := lineEntry{Nick: "foo", Text: "blah", Timestamp: ts}

	n := le.NiceTime()
	if n != "12:04:36" {
		t.Error("%s != %s", n, "12:04:36")
	}
}

func Test_lineEntryPermalink(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339Nano, "2015-02-15T12:04:36.439011141-05:00")
	le := lineEntry{Nick: "foo", Text: "blah", Timestamp: ts}

	p := le.Permalink()
	if p != "/logs/2015/02/15/#2015-02-15T12:04:36.439011141-05:00" {
		t.Error(p)
	}
}

type testcase struct {
	Input    string
	Expected string
}

func Test_normalizeNick(t *testing.T) {
	tests := []testcase{
		{"test", "test"},
		{"test_", "test"},
	}
	for _, test := range tests {
		if normalizeNick(test.Input) != test.Expected {
			t.Error("%s != %s", normalizeNick(test.Input), test.Expected)
		}
	}
}
