package main

import (
	"fmt"
	"html/template"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/oauth2"
	"google.golang.org/api/plus/v1"

	"google-musicmanager-go"
)

var scopes = []string{musicmanager.MusicManagerScope, plus.PlusMeScope}
var conf = googleMustConfigFromFile("credentials.json", scopes...)
var tpls = template.Must(template.New("static").
	Funcs(map[string]interface{}{"reverse": reversed}).
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
	/*
		http.Handle("/jobs/", &REST{
			Init:   initMusicManager,
			List:   jobsList,
			Delete: jobsCancel,
		})
	*/
}

func auth(_ interface{}, w http.ResponseWriter, r *http.Request) error {
	state, err := nonce(32)
	if err != nil {
		return err
	}
	// BUG(lor): Google Play Music Web Manager ignores the
	// "redirect_uris" property of the credentials file.  Instead,
	// a redirect URL is always created dynamically by appending
	// "/oauth2callback" to the scheme and host of the domain on
	// which Google Play Music Web Manager received the request for
	// "/auth".
	conf.RedirectURL = getRedirectURL(getContext(r))
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
	// Try to read all multipart files to memory.  This is necessary
	// due to the way tracksInsert works: the HTTP response is
	// returned when the upload starts, not when it finishes.
	// The upload is thus done in a goroutine that outlives the
	// request handler, and any temporary files it could create.
	//
	// We substract one from MaxInt64 because the code for
	// ParseMultipartForm adds one to the argument for whatever
	// reason.
	err := r.ParseMultipartForm(math.MaxInt64 - 1)
	if err != nil && err != http.ErrNotMultipart {
		return nil, err
	}
	// If either the access token or uploader ID cookie is missing,
	// autoredirect to the start of the authorization flow rather
	// than just report an error.  The redirect parameter lets the
	// user continue right where they left off.

	// BUG(lor): If the access_token cookie expires just before
	// submitting new tracks for upload, they will need to be
	// resubmitted after the auth flow has finished, as it cannot
	// preserve POST data.
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
	// Create and return a new Music Manager service.  On App
	// Engine, fixTransport turns off SSL verification for the
	// transport so that access to the android.clients.google.com
	// server works fine.
	c := getContext(r)
	client := conf.Client(c, &oauth2.Token{AccessToken: tok.Value})
	fixTransport(client.Transport.(*oauth2.Transport).Base)
	return musicmanager.New(client, id.Value)
}

func register(client interface{}, w http.ResponseWriter, r *http.Request) error {
	name := r.FormValue("name")
	if name == "" {
		name = "Google Play Music Web Manager"
	}
	err := client.(*musicmanager.Service).Register(name).Do()
	if err != nil {
		return err
	}
	if redirect := r.FormValue("redirect"); redirect != "" {
		http.Redirect(w, r, redirect, http.StatusFound)
	}
	fmt.Fprintln(w, musicmanager.GetRegisterError("OK"))
	return err
}

func tracksGet(client interface{}, w http.ResponseWriter, r *http.Request) error {
	ID := r.URL.Path
	track, err := client.(*musicmanager.Service).Tracks.Get(ID).Do()
	if err != nil {
		return err
	}
	if name := track.Name(); name != "" {
		// url.QueryEscape encodes spaces as plus signs, which
		// popular browsers don't understand.  Percent-encode
		// them instead.
		name = url.QueryEscape(name)
		if tmp, err := regexpReplaceAllString(`\+`, name, "%20"); err == nil {
			name = tmp
		}
		name = "attachment; filename*=UTF8-''" + name
		w.Header().Set("Content-Disposition", name)
	}
	if size := track.Size(); size > 0 {
		size := strconv.FormatInt(size, 10)
		w.Header().Set("Content-Length", size)
	}
	w.Header().Set("Content-Type", "audio/mpeg")
	_, err = io.Copy(w, track)
	return err
}

func tracksList(client interface{}, w http.ResponseWriter, r *http.Request) error {
	list := client.(*musicmanager.Service).Tracks.List()
	// Parse the options.
	//if tmp := list.UpdatedMin(r.FormValue("updatedMin")); tmp != nil {
	//	list = tmp
	//}
	purchasedOnly, err := strconv.ParseBool(r.FormValue("purchasedOnly"))
	if err == nil {
		list.PurchasedOnly(purchasedOnly)
	}
	if pageToken := r.FormValue("pageToken"); pageToken != "" {
		list.PageToken(pageToken)
	}
	// Execute the query.
	res, err := list.Do()
	if err != nil {
		return err
	}
	// Print the results.
	// GetTracksToExportResponse doesn't report back whether it
	// was obtained with *ExportType == ALL or
	// *ExportType == PURCHASED_AND_PROMOTIONAL, and I don't want
	// to define a new type just to pass this information to the
	// template.  Fortunately, the Go protobuf compiler includes
	// an XXX_unrecognized []byte field with every struct, which
	// is perfect for smuggling this information to the template.
	res.XXX_unrecognized = nil
	if purchasedOnly {
		res.XXX_unrecognized = []byte("true")
	}
	return tpls.ExecuteTemplate(w, "list.tpl", res)
}

func tracksInsert(client interface{}, w http.ResponseWriter, r *http.Request) error {
	f, _, err := r.FormFile("track")
	if err != nil {
		return err
	}
	defer f.Close()
	if err := checkSize(f); err != nil {
		return err
	}
	serverID, err := client.(*musicmanager.Service).Tracks.Insert(f).Do()
	if err != nil {
		return err
	}
	http.Redirect(w, r, "/tracks/", http.StatusFound)
	fmt.Fprintln(w, serverID)
	return nil
}

/*
func jobsList(client interface{}, w http.ResponseWriter, r *http.Request) error {
	jobs, err := client.(*musicmanager.Service).Jobs.List().Do()
	if err != nil {
		return err
	}
	for _, job := range jobs {
		fmt.Fprintf(w, "%+v\n", job)
	}
	return nil
}

func jobsCancel(client interface{}, w http.ResponseWriter, r *http.Request) error {
	err := client.(*musicmanager.Service).Jobs.Cancel().Do()
	if err != nil {
		return err
	}
	http.Redirect(w, r, "/jobs/", http.StatusSeeOther)
	return nil
}
*/
