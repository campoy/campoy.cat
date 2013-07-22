package app

import (
	"fmt"
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
)

var homeTmpl = template.Must(template.New("home").Parse(`
<html>
	<head>
		<title>{{.Title}}</title>
		<link rel="stylesheet" type="text/css" href="http://twitter.github.io/bootstrap/assets/css/bootstrap.css">
	</head>

	<body>
		<div class="hero-unit">
			<h1>{{.Title}}</h1>
			<p>{{.Message}}</p>
			</p>
			<ul>
				{{range .Links}}
					<li><a href="{{.URL}}">{{.Name}}</a></li>
				{{end}}		
			</ul>
		</div>
	</body>
</html>
`))

func init() {
	http.HandleFunc("/", errorHandler(homeHandler))
}

func errorHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			http.Error(w, "Ooops! something bad happened", http.StatusInternalServerError)
			appengine.NewContext(r).Errorf("handling %q: %v", r.URL.Path, err)
			return
		}
	}
}

type Link struct {
	URL  string
	Path string
	Name string
}

type Page struct {
	Title   string
	Message string
	Links   []Link `datastore:-`
}

func page(c appengine.Context) (*Page, error) {
	p := Page{}

	k := datastore.NewKey(c, "page", "page", 0, nil)
	err := datastore.Get(c, k, &p)
	if err != nil {
		return nil, err
	}

	_, err = datastore.NewQuery("link").GetAll(c, &p.Links)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func redirect(c appengine.Context, w http.ResponseWriter, r *http.Request) error {
	link := Link{}
	path := r.URL.Path[1:]
	it := datastore.NewQuery("link").Filter("Path=", path).Run(c)
	_, err := it.Next(&link)
	if err != nil {
		return fmt.Errorf("finding link %q: %v", path, err)
	}

	http.Redirect(w, r, link.URL, http.StatusMovedPermanently)
	return nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	p, err := page(c)
	if err != nil {
		return fmt.Errorf("getting page: %v", err)
	}

	if r.URL.Path == "/" {
		err := homeTmpl.Execute(w, p)
		if err != nil {
			return fmt.Errorf("rendering page: %v", err)
		}
		return nil
	}

	return redirect(c, w, r)
}
