package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"

	_ "github.com/motemen/go-loghttp/global"
)

const baseApiUrl = "https://api.airtable.com/"

type transport struct {
	apiKey string
}

func(t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	log.Printf("Sending request to %q\n", r.URL.String())
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))
	r.Header.Del("X-Forwarded-For")
	return http.DefaultTransport.RoundTrip(r)
}

func newProxyHandler(apiKey string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uri, err := url.Parse(baseApiUrl)
		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(uri)
		proxy.Transport = &transport{apiKey: apiKey}

		proxy.ServeHTTP(w, r)
	}
}

func main() {
	airTableApiKey := os.Getenv("AIR_TABLE_API_KEY")

	router := mux.NewRouter()
	router.HandleFunc("/{version}/{app}/{resource}", newProxyHandler(airTableApiKey)).Methods("POST")

	srv := &http.Server{
		Handler:      router,
		Addr:         ":4242",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
