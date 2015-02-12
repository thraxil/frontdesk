package main

import (
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

func yearView(w http.ResponseWriter, r *http.Request, s *site, year string) {

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

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// just ignore this crap
}
