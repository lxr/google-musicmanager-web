package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/oauth2"

	"github.com/dhowden/tag"
	"google.golang.org/api/plus/v1"

	"my-git.appspot.com/go.google.musicmanager"
)

var funcMap = map[string]interface{}{
	"incr": func(i int) int { return i + 1 },
	"time": unix2Time,
}
var scopes = []string{musicmanager.Scope, plus.PlusMeScope}
var conf = googleMustConfigFromFile("credentials.json", scopes...)
var tpls = template.Must(template.New("static").
	Funcs(funcMap).
	ParseGlob("static/*.tpl"))

func init() {
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.Handle("/auth", &REST{Get: auth})
	http.Handle("/oauth2callback", &REST{Get: oauth2callback})
	http.Handle("/register", &REST{
		Init: initMusicManager,
		Get:  register,
	})
	http.Handle("/tracks/", &REST{
		Init:   initMusicManager,
		Get:    tracksGet,
		List:   tracksList,
		Insert: tracksInsert,
	})
}

func auth(_ interface{}, w http.ResponseWriter, r *http.Request) error {
	state, err := nonce(32)
	if err != nil {
		return err
	}
	conf.RedirectURL = getRedirectURL(r)
	httpSetCookie(w, r, "state", state)
	httpSetCookie(w, r, "redirect", r.FormValue("redirect"))
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
	return nil
}

func oauth2callback(_ interface{}, w http.ResponseWriter, r *http.Request) error {
	// Confirm that the state matches the nonce we stored
	// (See https://tools.ietf.org/html/rfc6749#section-10.12.)
	rstate := r.FormValue("state")
	astate, err := r.Cookie("state")
	if err != nil || rstate != astate.Value {
		return &RESTError{
			Code: http.StatusBadRequest,
			Message: "state parameter and cookie mismatch" +
				"; have you perhaps disabled cookies?",
		}
	}
	// Exchange the authorization code for an access token.
	c := getContext(r)
	tok, err := conf.Exchange(c, r.FormValue("code"))
	if err != nil {
		return err
	}
	// BUG(lor): Google Play Music does not allow downloading tracks
	// with an uploader_id that is not sufficiently
	// "MAC address-like" (perhaps it only checks for a colon?)
	// The /oauth2callback endpoint generates the uploader_id by
	// injecting a colon between every two digits of the user's
	// Google Account ID, which appears to suffice.
	client := conf.Client(c, tok)
	plus, err := plus.New(client)
	if err != nil {
		return err
	}
	person, err := plus.People.Get("me").Do()
	if err != nil {
		return err
	}
	id, err := regexpReplaceAllString(`(..)`, person.Id, "$1:")
	if err != nil {
		return err
	}
	// Wipe the state and redirect cookies, no longer necessary,
	// and store the access token and uploader ID as cookies.
	httpSetCookie(w, r, "state", "")
	httpSetCookie(w, r, "redirect", "")
	httpSetCookie(w, r, "access_token", tok.AccessToken)
	httpSetCookie(w, r, "uploader_id", id)
	// We autoredirect to the registration endpoint for convenience.
	// If the redirect cookie was provided, remember to tell the
	// registration endpoint to continue there after success.
	registerURL := "/register"
	redirect, err := r.Cookie("redirect")
	if err == nil && redirect.Value != "" {
		registerURL += "?redirect=" + url.QueryEscape(redirect.Value)
	}
	http.Redirect(w, r, registerURL, http.StatusFound)
	return nil
}

func initMusicManager(r *http.Request) (interface{}, error) {
	// If either the access token or uploader ID cookie is missing,
	// autoredirect to the start of the authorization flow rather
	// than just report an error.  The redirect parameter lets the
	// user continue right where they left off.

	// BUG(lor): If the access_token cookie expires just before
	// uploading a new track, the track will need to be resubmitted
	// after the auth flow has finished, as the flow cannot preserve
	// POST data.
	tok, _ := r.Cookie("access_token")
	id, _ := r.Cookie("uploader_id")
	if tok == nil || id == nil {
		path := url.QueryEscape(r.URL.Path + "?" + r.URL.RawQuery)
		return nil, &RESTError{
			Code:     http.StatusFound,
			Message:  "missing credentials",
			Location: "/auth?redirect=" + path,
		}
	}
	// Create and return a new Music Manager service.
	c := getContext(r)
	client := conf.Client(c, &oauth2.Token{AccessToken: tok.Value})
	client.Transport.(*oauth2.Transport).Base = getTransport(c)
	return musicmanager.NewClient(client, id.Value)
}

func register(client interface{}, w http.ResponseWriter, r *http.Request) error {
	name := r.FormValue("name")
	if name == "" {
		name = "Google Play Music Web Manager"
	}
	err := client.(*musicmanager.Client).Register(name)
	if err != nil {
		return err
	}
	if redirect := r.FormValue("redirect"); redirect != "" {
		http.Redirect(w, r, redirect, http.StatusFound)
	}
	fmt.Fprintln(w, "registration successful")
	return nil
}

func tracksGet(client interface{}, w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Path
	url, err := client.(*musicmanager.Client).ExportTrack(id)
	if err != nil {
		return err
	}
	http.Redirect(w, r, url, http.StatusFound)
	return nil
}

func tracksList(client interface{}, w http.ResponseWriter, r *http.Request) error {
	updatedMin, _ := time.Parse(time.RFC3339Nano, r.FormValue("updatedMin"))
	purchasedOnly, _ := strconv.ParseBool(r.FormValue("purchasedOnly"))
	continuationToken := r.FormValue("pageToken")
	trackList, err := client.(*musicmanager.Client).ListTracks(
		purchasedOnly,
		updatedMin.UnixNano()/1000,
		continuationToken,
	)
	if err != nil {
		return err
	}
	return tpls.ExecuteTemplate(w, "list.tpl", trackList)
}

func tracksInsert(client interface{}, w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	f, fh, err := r.FormFile("track")
	if err != nil {
		return err
	}
	defer f.Close()
	track, err := parseTrack(f, fh.Filename)
	if err != nil {
		return err
	}
	urls, errs := client.(*musicmanager.Client).ImportTracks([]*musicmanager.Track{track})
	if errs[0] != nil {
		return errs[0]
	}
	cl := &http.Client{Transport: getTransport(getContext(r))}
	resp, err := cl.Post(urls[0], "audio/mpeg", f)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	http.Redirect(w, r, r.RequestURI, http.StatusFound)
	io.Copy(w, resp.Body) // for debugging
	return nil
}

func parseTrack(r io.ReadSeeker, name string) (*musicmanager.Track, error) {
	sum, err := tag.Sum(r)
	err = rewind(r, err)
	if err != nil {
		return nil, err
	}
	metadata, err := tag.ReadFrom(r)
	err = rewind(r, err)
	switch {
	case err == tag.ErrNoTagsFound:
		return &musicmanager.Track{
			ClientId: sum,
			Title:    name,
		}, nil
	case err != nil:
		return nil, err
	}
	ti, tn := metadata.Track()
	di, dn := metadata.Disc()
	return &musicmanager.Track{
		ClientId:        sum,
		Title:           metadata.Title(),
		Album:           metadata.Album(),
		Artist:          metadata.Artist(),
		AlbumArtist:     metadata.AlbumArtist(),
		Composer:        metadata.Composer(),
		Year:            metadata.Year(),
		Genre:           metadata.Genre(),
		TrackNumber:     ti,
		TotalTrackCount: tn,
		DiscNumber:      di,
		TotalDiscCount:  dn,
	}, nil
}
