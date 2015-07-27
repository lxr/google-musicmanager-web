package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// googleMustConfigFromFile reads an *oauth2.Config from the given JSON
// file an associates it with the given scopes, or panics trying.
func googleMustConfigFromFile(filename string, scope ...string) *oauth2.Config {
	jsonKey, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	conf, err := google.ConfigFromJSON(jsonKey, scope...)
	if err != nil {
		panic(err)
	}
	return conf
}

// httpSetCookie sets an HTTP-only cookie for the root of the domain,
// secure or not depending on whether the request was received over TLS.
// Setting a cookie to the empty string deletes it.
func httpSetCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	// All Google OAuth 2.0 tokens I've seen have had this as their
	// expiry value, so let's mimic that behavior.
	maxage := 3598
	// BUG(lor): There is a race condition in that an OAuth 2.0
	// token may expire before the access_token cookie in which it
	// is stored, so that a REST method call results in an error
	// response and not an automatic redirect.
	if value == "" {
		maxage = -1
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxage,
		Secure:   r.TLS != nil,
		HttpOnly: true,
	})
}

// getRedirectURL generates an OAuth2 redirect URL redirecting to the
// path /oauth2callback on the host on which the request was received.
func getRedirectURL(r *http.Request) string {
	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}
	return scheme + r.Host + "/oauth2callback"
}

// nonce returns a base64-encoded string of n random bytes.
func nonce(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// regexpReplaceAllString replaces all occurrences of pattern in src
// with repl and returns the result, or an error if the pattern couldn't
// be compiled.
func regexpReplaceAllString(pattern, src, repl string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(src, repl), nil
}

// unix2Time returns the local Time corresponding to the given Unix time
// in microseconds.
func unix2Time(msec int64) time.Time {
	return time.Unix(msec/1000000, (msec%1000000)*1000)
}

// fakeReadSeeker wraps an io.Reader in a buffer such that it can be
// transparently seeked.
type fakeReadSeeker struct {
	i      int
	buf    []byte
	source io.Reader
}

func newFakeReadSeeker(r io.Reader) io.ReadSeeker {
	return &fakeReadSeeker{source: r}
}

func (rs *fakeReadSeeker) read(n int) (int, error) {
	p := make([]byte, n)
	n, err := rs.source.Read(p)
	rs.buf = append(rs.buf, p[:n]...)
	return n, err
}

func (rs *fakeReadSeeker) Read(p []byte) (n int, err error) {
	if want := len(p) - len(rs.buf[rs.i:]); want > 0 {
		_, err = rs.read(want)
	}
	n = copy(p, rs.buf[rs.i:])
	rs.i += n
	return n, err
}

func (rs *fakeReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case 0:
		abs = offset
	case 1:
		abs = int64(rs.i) + offset
	case 2:
		b, err := ioutil.ReadAll(rs.source)
		rs.buf = append(rs.buf, b...)
		if err != nil {
			return 0, err
		}
		abs = int64(len(rs.buf)) + offset
	default:
		return 0, errors.New("unknown whence value to seek")
	}
	if abs < 0 {
		return 0, errors.New("seek to negative offset")
	}
	if want := int(abs) - len(rs.buf); want > 0 {
		_, err := rs.read(want)
		if err != nil && err != io.EOF {
			return 0, err
		}
	}
	if max := int64(len(rs.buf)); abs > max {
		abs = max
	}
	rs.i = int(abs)
	return abs, nil
}
