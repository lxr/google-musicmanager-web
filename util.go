package main

import (
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"

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

// reversed returns a reversed copy of the slice, array or string
// argument.
func reversed(x interface{}) []interface{} {
	s := reflect.ValueOf(x)
	n := s.Len()
	r := make([]interface{}, n)
	for i := 0; i < n; i++ {
		r[n-i-1] = s.Index(i).Interface()
	}
	return r
}
