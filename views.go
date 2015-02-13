package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

type indexPage struct {
	Title string
	Years []string
}

func indexHandler(w http.ResponseWriter, r *http.Request, s *site) {
	p := indexPage{
		Title: "front desk",
		Years: s.years(),
	}
	t, _ := template.New("index").Parse(indexTemplate)
	t.Execute(w, p)
}

type linksPage struct {
	Title string
	Links []linkEntry
}

func linksHandler(w http.ResponseWriter, r *http.Request, s *site) {
	recentLinks := s.recentLinks()
	p := linksPage{
		Title: "front desk: links",
		Links: recentLinks,
	}
	t, _ := template.New("links").Parse(linksTemplate)
	t.Execute(w, p)
}

func logsHandler(w http.ResponseWriter, r *http.Request, s *site) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) == 4 {
		yearView(w, r, s, parts[2])
		return
	}
	if len(parts) == 5 {
		monthView(w, r, s, parts[2], parts[3])
		return
	}
	if len(parts) == 6 {
		dayView(w, r, s, parts[2], parts[3], parts[4])
		return
	}
	http.Error(w, "bad request", 400)
}

type yearPage struct {
	Title  string
	Year   string
	Months []string
}

func yearView(w http.ResponseWriter, r *http.Request, s *site, year string) {
	p := yearPage{
		Title:  year,
		Year:   year,
		Months: s.monthsForYear(year),
	}
	t, _ := template.New("year").Parse(yearTemplate)
	t.Execute(w, p)
}

type monthPage struct {
	Title string
	Year  string
	Month string
	Days  []string
}

func monthView(w http.ResponseWriter, r *http.Request, s *site, year, month string) {
	p := monthPage{
		Title: fmt.Sprintf("%s-%s", year, month),
		Year:  year,
		Month: month,
		Days:  s.daysForMonth(year, month),
	}
	t, _ := template.New("month").Parse(monthTemplate)
	t.Execute(w, p)
}

type dayPage struct {
	Title string
	Year  string
	Month string
	Day   string
	Lines []lineEntry
}

func dayView(w http.ResponseWriter, r *http.Request, s *site, year, month, day string) {
	p := dayPage{
		Title: fmt.Sprintf("%s-%s-%s", year, month, day),
		Year:  year,
		Month: month,
		Day:   day,
		Lines: s.linesForDay(year, month, day),
	}
	t, _ := template.New("day").Parse(dayTemplate)
	t.Execute(w, p)
}

type smoketestResponse struct {
	Status       string   `json:"status"`
	TestClasses  int      `json:"test_classes"`
	TestsRun     int      `json:"tests_run"`
	TestsPassed  int      `json:"tests_passed"`
	TestsFailed  int      `json:"tests_failed"`
	TestsErrored int      `json:"tests_errored"`
	Time         float64  `json:"time"`
	ErroredTests []string `json:"errored_tests"`
	FailedTests  []string `json:"failed_tests"`
}

func smoketestHandler(w http.ResponseWriter, r *http.Request, s *site) {
	var status string
	var tests int
	if backoff == 0 {
		status = "PASS"
		tests = 1
	} else {
		status = "FAIL"
		tests = 0
	}
	sr := smoketestResponse{
		Status:       status,
		TestClasses:  1,
		TestsRun:     1,
		TestsPassed:  tests,
		TestsFailed:  1 - tests,
		TestsErrored: 0,
		Time:         1.0,
	}

	h := r.Header.Get("Accept")
	if strings.Index(h, "application/json") != -1 {
		b, _ := json.Marshal(sr)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return
	}
	smokeTemplate := `{{.Status}}
test classes: 1
tests run: 1
tests passed: {{.TestsPassed}}
tests failed: {{.TestsFailed}}
tests errored: 0
time: 1.0ms
`
	t, _ := template.New("smoketest").Parse(smokeTemplate)
	w.Header().Set("Content-Type", "text/plain")
	t.Execute(w, sr)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// just ignore this crap
}
