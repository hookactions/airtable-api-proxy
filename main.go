package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const baseApiUrl = "https://api.airtable.com/"

type transport struct {
	apiKey string
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	log.Printf("Sending request, %s %s", r.Method, r.URL.String())

	r.Host = r.URL.Host
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))
	r.Header.Del("X-Forwarded-For")

	return http.DefaultTransport.RoundTrip(r)
}

func newProxyHandler(apiKey string) func(http.ResponseWriter, *http.Request) {
	uri, err := url.Parse(baseApiUrl)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(uri)
		proxy.Transport = &transport{apiKey: apiKey}

		proxy.ServeHTTP(w, r)
	}
}

func main() {
	port := flag.Int("port", 4242, "port to run the server on")
	flag.Parse()

	airTableApiKey := os.Getenv("AIR_TABLE_API_KEY")

	router := mux.NewRouter()
	router.HandleFunc("/{version}/{app}/{resource}", newProxyHandler(airTableApiKey)).Methods("POST", "OPTIONS")

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", *port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Listening on %s ...", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
