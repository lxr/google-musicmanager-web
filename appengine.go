// +build appengine

package main

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// BUG(lor): The App Engine URL Fetch API has an upload limit of 10
// megabytes (see https://cloud.google.com/appengine/docs/go/urlfetch/).
const MaxUploadSize int64 = 10 << 20

func getContext(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func getTransport(ctx context.Context) http.RoundTripper {
	// The default urlfetch deadline of 5 seconds is way too short
	// for uploading music.  The maximum is one minute.
	ctx, _ = context.WithTimeout(ctx, time.Minute)
	return &urlfetch.Transport{
		Context: ctx,
		// BUG(lor): The App Engine proxy appears not to perform
		// Server Name Indication (see https://tools.ietf.org/html/rfc6066#section-3),
		// which is required to access the server
		// android.clients.google.com over TLS.  We turn off TLS
		// verification instead.
		AllowInvalidServerCertificate: true,
	}
}
