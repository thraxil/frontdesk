package main

var indexTemplate = `
<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" />
</head>
<body>
<div class="container">
<ol class="breadcrumb">
  <li class="active">Home</li>
</ol>

<h1>{{.Title}}</h1>

<div class="list-group">
<a class="list-group-item" href="/links/">Recent Links</a>
<a class="list-group-item" href="/search/">Search</a>
{{ range .Years }}
<a class="list-group-item" href="/logs/{{ . }}/">Full Chat Logs {{ . }}</a>
{{ end }}
</div>
</div>
</html>

`

var dayTemplate = `
<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" />
<script src="//code.jquery.com/jquery-1.11.2.min.js"></script>
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
<tr id="{{.Key}}">
  <td><a name="{{.Key}}"></a><a href="#{{.Key}}">{{.NiceTime}}</a></td>
  <td>&lt;<b>{{.Nick}}</b>&gt;</td>
  <td><tt>{{.Text}}</tt></td>
</tr>
{{ end }}
</table>
</div>
<script>
$(document).ready ( function () {
	var id = location.hash.substr(1);
	if (id) {
     $('tr').each(function (k, e) {
        if (e.id === id) {
           $(e).addClass('success');
        }
     });
  }
});
</script>
</html>

`

var searchTemplate = `
<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" />
<script src="//code.jquery.com/jquery-1.11.2.min.js"></script>
</head>
<body>
<div class="container">
<ol class="breadcrumb">
  <li><a href="/">Home</a></li>
  <li class="active">Search</li>
</ol>
<h1>{{.Title}}</h1>

<form action="." method="get" class="form-inline">
<div class="form-group">
<input type="text" name="q" value="{{.Query}}" class="form-control"/>
<input type="submit" value="search" class="btn btn-primary" />
</div>
</form>
<p>{{.Results.Total}} Hits</p>
<table class="table table-striped table-condensed">
{{ range .Lines }}
<tr>
  <td><a href="{{.Permalink}}">{{.Timestamp.Month}} {{.Timestamp.Day}} {{.Timestamp.Year}}  {{.NiceTime}}</a></td>
  <td>&lt;<b>{{.Nick}}</b>&gt;</td>
  <td><tt>{{.Text}}</tt></td>
</tr>

{{ end }}
</table>

</div>
</html>
`

var emptySearchTemplate = `
<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" />
<script src="//code.jquery.com/jquery-1.11.2.min.js"></script>
</head>
<body>
<div class="container">
<ol class="breadcrumb">
  <li><a href="/">Home</a></li>
  <li class="active">Search</li>
</ol>
<h1>{{.Title}}</h1>

<form action="." method="get" class="form-inline">
<div class="form-group">
<input type="text" name="q" value="" class="form-control"/>
<input type="submit" value="search" class="btn btn-primary" />
</div>
</form>
</div>
</body>
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

var linksTemplate = `
<html>
<head>
<title>{{.Title}}</title>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" />
</head>
<body>
<div class="container">
<ol class="breadcrumb">
  <li><a href="/">Home</a></li>
  <li class="active">Recent Links</li>
</ol>
<h1>{{.Title}}</h1>
<table class="table table-striped table-condensed">
{{ range .Links }}
<tr>
  <td><a href="{{.URL}}">{{.Title}}</a></td>
  <td><b>{{.Nick}}</b></td>
  <td>{{.FormattedTimestamp}}<td>
  <td><a href="{{.DiscussionLink}}">discussion</a></td>
</tr>
{{ end }}
</table>
</div>
</html>

`
