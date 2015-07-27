// +build !appengine

package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// Default address to serve on.
var addr = "localhost:8080"

func getContext(r *http.Request) context.Context {
	return oauth2.NoContext
}

func getRedirectURL(ctx context.Context) string {
	// BUG(lor): The standalone webserver cannot serve over HTTPS.
	// Even if an HTTPS bridge is configured for it, the OAuth 2.0
	// redirect URL is given as HTTP.
	return "http://" + addr + "/oauth2callback"
}

func fixTransport(transport http.RoundTripper) {}

func main() {
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
