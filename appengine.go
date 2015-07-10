// +build appengine

package main

import (
	"errors"
	"io"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func getContext(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func getRedirectURL(ctx context.Context) string {
	scheme := "https://"
	// We assume the dev app server is never accessed over HTTPS.
	if appengine.IsDevAppServer() {
		scheme = "http://"
	}
	return scheme + appengine.DefaultVersionHostname(ctx) + "/oauth2callback"
}

func checkSize(s io.Seeker) error {
	size, err := s.Seek(0, 2)
	if err != nil {
		return err
	}
	// BUG(lor): The App Engine URL Fetch API has a 10 MB limit on
	// request size (https://cloud.google.com/appengine/docs/go/urlfetch/#Go_Quotas_and_limits),
	// so files larger than that cannot be uploaded on it.

	// You'd think that trying to send over 10 MB inside App Engine
	// would return an error or something, but apparently it just
	// throws an exception in the Python wrapper, at least on the
	// dev server... so validating the size is the client code's
	// responsibility.
	if size > 10<<20 {
		return errors.New("urlfetch: file too large")
	}
	_, err = s.Seek(0, 0)
	return err
}

func fixTransport(transport http.RoundTripper) {
	// BUG(lor): The App Engine proxy lacks a valid certificate for
	// https://android.clients.google.com, which is used by most
	// endpoints, so we need to turn off SSL verification to access
	// it on the App Engine.
	t := transport.(*urlfetch.Transport)
	t.AllowInvalidServerCertificate = true
}
