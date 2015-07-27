// +build appengine

package main

import (
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func getContext(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func fixTransport(transport http.RoundTripper) {
	// BUG(lor): The App Engine proxy appears not to perform Server
	// Name Indication (see https://tools.ietf.org/html/rfc6066#section-3),
	// which is required to access the server android.clients.google.com
	// over TLS.  We turn off TLS verification instead.
	t := transport.(*urlfetch.Transport)
	t.AllowInvalidServerCertificate = true
}
