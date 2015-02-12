package main

var indexTemplate = `
<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" />
</head>
<body>
<div class="container">
<h1>{{.Title}}</h1>
<table class="table">
{{ range .Years }}
<tr><th><a href="/logs/{{ . }}/">{{ . }}</a></th></tr>
{{ end }}
</table>
</div>
</html>

`

var dayTemplate = `
<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" />
</head>
<body>
<div class="container">
<ol class="breadcrumb">
  <li><a href="/">Home</a></li>
  <li><a href="/logs/{{.Year}}/">{{.Year}}</a></li>
  <li><a href="/logs/{{.Year}}/{{.Month}}/">{{.Month}}</a></li>
  <li class="active">{{.Day}}</li>
</ol>
<h1>{{.Title}}</h1>
<table class="table table-striped table-condensed">
{{ range .Lines }}
<tr>
  <td><a name="{{.Key}}"></a><a href="#{{.Key}}">{{.NiceTime}}</a></td>
  <td><b>{{.Nick}}</b></td>
  <td><tt>{{.Text}}</tt></td>
</tr>
{{ end }}
</table>
</div>
</html>

`

var monthTemplate = `
<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" />
</head>
<body>
<div class="container">
<ol class="breadcrumb">
  <li><a href="/">Home</a></li>
  <li><a href="/logs/{{.Year}}/">{{.Year}}</a></li>
  <li class="active">{{.Month}}</li>
</ol>
<h1>{{.Title}}</h1>
<table class="table table-striped table-condensed">
{{ range .Days }}
<tr><td><a href="{{.}}/">{{.}}</a></td></tr>
{{ end }}
</table>
</div>
</html>`

var yearTemplate = `
<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" />
</head>
<body>
<div class="container">
<ol class="breadcrumb">
  <li><a href="/">Home</a></li>
  <li class="active">{{.Year}}</li>
</ol>
<h1>{{.Title}}</h1>
<table class="table table-striped table-condensed">
{{ range .Months }}
<tr><td><a href="{{.}}/">{{.}}</a></td></tr>
{{ end }}
</table>
</div>
</html>`
