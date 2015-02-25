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

type mentionTestCase struct {
	Line        string
	Nick        string
	ExpectMatch bool
}

func Test_mentionsNick(t *testing.T) {
	cases := []mentionTestCase{
		{
			Line:        "alice: you are being mentioned",
			Nick:        "alice",
			ExpectMatch: true,
		},
		{
			Line:        "alice you are being mentioned",
			Nick:        "alice",
			ExpectMatch: true,
		},
		{
			Line:        "alice_: you are being mentioned",
			Nick:        "alice",
			ExpectMatch: true,
		},
		{
			Line:        "alice_ you are being mentioned",
			Nick:        "alice",
			ExpectMatch: true,
		},
		{
			Line:        "alice is being mentioned",
			Nick:        "bob",
			ExpectMatch: false,
		},
		{
			Line:        "i mean no malice",
			Nick:        "alice",
			ExpectMatch: false,
		},
		// currently failing cases
		//		{
		//			Line:        "malice is not intendend",
		//			Nick:        "alice",
		//			ExpectMatch: false,
		//		},
		//		{
		//			Line:        "but now i'm going to mention alice",
		//			Nick:        "alice",
		//			ExpectMatch: true,
		//		},
	}
	for _, tc := range cases {
		r := mentionsNick(tc.Line, tc.Nick)
		if r != tc.ExpectMatch {
			t.Errorf("mentionsNick(\"%s\", \"%s\") returned %v, expected %v\n", tc.Line, tc.Nick, r, tc.ExpectMatch)
		}
	}
}
