package main

import (
	"testing"
	"time"
)

func Test_mentionPermalink(t *testing.T) {
	m := mention{
		"foo",
		2015, 02, 17, "a-key", "some text", time.Now(),
	}
	if m.Permalink() != "/logs/2015/02/17/#a-key" {
		t.Error("bad permalink")
	}
}
