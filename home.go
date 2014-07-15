package app

import (
	"fmt"
	"net/http"

	"appengine"
	"appengine/datastore"
)

func init() {
	http.HandleFunc("/", errorHandler(homeHandler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) error {
	var link struct {
		URL  string
		Name string
		Path string
	}
	path := r.URL.Path[1:]
	c := appengine.NewContext(r)
	it := datastore.NewQuery("link").Filter("Path=", path).Run(c)
	_, err := it.Next(&link)
	if err != nil {
		return fmt.Errorf("finding link %q: %v", path, err)
	}

	http.Redirect(w, r, link.URL, http.StatusMovedPermanently)
	return nil
}

func errorHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			appengine.NewContext(r).Errorf("handling %q: %v", r.URL.Path, err)
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}
	}
}
