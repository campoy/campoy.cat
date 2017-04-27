package app

import (
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
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
	ctx := appengine.NewContext(r)
	path := strings.TrimPrefix(r.URL.Path[1:], "l/")
	it := datastore.NewQuery("link").Filter("Path=", path).Run(ctx)
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
			ctx := appengine.NewContext(r)
			log.Errorf(ctx, "handling %q: %v", r.URL.Path, err)
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}
	}
}
