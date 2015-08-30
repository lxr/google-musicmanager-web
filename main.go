// +build !appengine

package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const MaxUploadSize = int64(^uint64(0) >> 1) // max int64

func getContext(r *http.Request) context.Context {
	return oauth2.NoContext
}

func getTransport(ctx context.Context) http.RoundTripper {
	return http.DefaultTransport
}

func main() {
	addr := "localhost:8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
