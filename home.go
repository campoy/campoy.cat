package app

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

var homeTmpl = template.Must(template.ParseFiles("home.html"))

func init() {
	http.HandleFunc("/", errorHandler(homeHandler))
	http.HandleFunc("/login", logHandler(user.LoginURL))
	http.HandleFunc("/logout", logHandler(user.LogoutURL))
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
	Brand   string
	Title   string
	Message template.HTML
	Links   []Link `datastore:-`
}

func page(c appengine.Context, lan string) (*Page, error) {
	p := Page{}

	kl := lan
	if len(kl) == 0 {
		kl = "page"
	}

	k := datastore.NewKey(c, "page", kl, 0, nil)
	err := datastore.Get(c, k, &p)
	if err == datastore.ErrNoSuchEntity && user.Current(c) != nil && user.Current(c).Admin {
		if len(lan) == 0 {
			p.Message = "go configurate your page in the datastore visualizer"
		} else {
			p.Message = template.HTML("page for " + lan + " is no defined.")
		}
		_, err = datastore.Put(c, k, &p)
	}
	if err != nil {
		return nil, err
	}

	_, err = datastore.NewQuery("link").Order("Name").GetAll(c, &p.Links)
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

	lan := r.FormValue("l")
	if lan == "" {
		lan = r.Header.Get("Accept-Language")
		lan = strings.SplitN(lan, "-", 2)[0]
	}

	p, err := page(c, lan)
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

func logHandler(get func(appengine.Context, string) (string, error)) http.HandlerFunc {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		c := appengine.NewContext(r)

		url, err := get(c, "/")
		if err != nil {
			return err
		}
		http.Redirect(w, r, url, http.StatusMovedPermanently)
		return nil
	})
}
