// +build !appengine

package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func getContext(r *http.Request) context.Context {
	return oauth2.NoContext
}

func fixTransport(transport http.RoundTripper) {}

func main() {
	addr := "localhost:8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
