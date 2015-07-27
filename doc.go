/*

Google Play Music Web Manager presents a simple RESTful web interface
for managing your Google Play Music library. It takes a single argument,
the TCP address where to start serving (default localhost:8080.) The
methods and endpoints it supports are as follows:

	GET /tracks/
		Returns an HTML listing of all tracks in the user's
		library, from most recently accessed to the least.
		Takes the optional query parameters purchasedOnly=true
		to filter only for purchased and promotional tracks,
		updatedMin={{RFC3339}} to filter out tracks that were
		last modified before the given timestamp, and
		pageToken={string} to page through large result sets.
		(The token for the next page can be found in the
		data-token attribute of the a[rel=next] element, if
		one exists.)

		In the listing, the id attribute of a table row contains
		the ID of the corresponding track.

	GET /tracks/{id}
		Redirects the client to a temporary
		sj.googleusercontent.com download URL for the given
		track.  The URL is also contained in the redirect
		response body.

	POST /tracks/
		Redirects the client to a temporary
		uploadsj.clients.google.com upload URL.  The URL is also
		contained in the redirect response body.  The redirect
		is performed with the 307 Temporary Redirect status
		code, which should transparently preserve the POST data
		in modern clients.  The POST body should be raw MP3
		data.  Google Play Music Web Manager only needs to read
		the file metadata in order to generate the upload URL,
		so ID3v2 tags are recommended, as the response handler
		can then return without reading the entire file.  (A
		consequence of this is that a Google Play Music Web
		Manager needs to read a file with no tags in its
		entirety, as any ID3v1 tags are located after the end of
		the audio data.)

All endpoints expect the request to carry the access_token and
uploader_id cookies, which are a Google OAuth 2.0 token carrying the
scope https://www.googleapis.com/auth/musicmanager and the ID of an
authorized device under your Play Music account (see
https://play.google.com/music/listen#/accountsettings, and inspect the
id attribute of the div.device-list-item elements) respectively.
For convenience, Google Play Music Web Manager provides the following
endpoints for acquiring these cookies:

	GET /auth
		Redirects the user to a Google consent page asking the
		user to authorize Google Play Music Web Manager with
		their account.  This page asks for the additional scope
		https://www.googleapis.com/auth/plus.me in order to
		discover the user's Google Account ID, which is used
		to generate the uploader_id cookie.  If authorization
		has been previously granted, this page redirects
		directly back to Google Play Music Web Manager (assuming
		all proper Google cookies are present for the whole
		transaction, i.e. a browser environment.)
		This endpoint takes the optional query parameter
		redirect={path}, which can be used to autoredirect to
		the desired page when the auth flow is complete.

	GET /oauth2callback
		This should not be directly called by the user; the
		Google consent page redirects to it once authorization
		has been granted or denied.  If authorization was
		granted, it computes and stores the uploader_id
		and access_token cookies with appropriate expiration
		times and redirects to the next page in the auth flow,
		passing on the redirect query parameter if given to
		/auth.

	GET /register
		Authorizes uploader_id as a device under the user's
		Play Music account and with the name "Google Play Music
		Web Manager".  The optional query parameter
		name={string} can be used to change this.  The
		query parameter redirect={path} can be provided to
		autoredirect to the given page on success; otherwise,
		this endpoint simply reports success or failure.

		Remember that there is a limit to how many devices an
		account can have, how many devices one can deauthorize
		in a year, and with how many accounts a device can be
		authorized, so be careful with the uploader_id cookies
		you pass to this endpoint.

The main endpoints redirect to the start of the auth flow if either
access_token or uploader_id is missing, and use the redirect query
parameter to return back once the flow is completed, so using the
interface through a web browser is a relatively seamless experience.

Google Play Music Web Manager can be run either as a standalone web
server or on Google App Engine, though the latter comes with a few
gotchas.  See the source file appengine.go for details.

On startup, Google Play Music Web Manager expects to find in the current
working directory a file called credentials.json, which contains
web application credentials for a Google Cloud project with the Google+
API enabled, and a Go html/template file static/list.tpl. The latter
is included with the default distribution; the former can be acquired by
registering a new project in the Google Developers Console
(https://console.developers.google.com) and creating a new client ID for
a web application under the "Credentials" tab.  Note that Google Play
Music Web Manager ignores the redirect_uris key of the credentials.json
file; the OAuth2 redirect URL is instead generated dynamically by
appending "/oauth2callback" to the host on which the request was
received.

*/
package main

// BUG(lor): Embedded album art is not currently uploaded.

// BUG(lor): Until someone writes a pure-Go audio transcoding library,
// Google Play Music Web Manager cannot perform song matching, and it
// can only support MP3 files.

// BUG(lor): Use of this program may constitute a violation of Google's
// terms of service under paragraph 2 of
// https://www.google.com/intl/en/policies/terms/#toc-services.
