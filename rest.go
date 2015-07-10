package main

import (
	"net/http"
	"strings"
)

var RESTMethodNotAllowed = &RESTError{
	Code:    http.StatusMethodNotAllowed,
	Message: "405 method not allowed",
}

type RESTError struct {
	Code     int
	Message  string
	Location string
}

func (e *RESTError) WriteTo(w http.ResponseWriter) {
	if e.Location != "" {
		w.Header().Set("Location", e.Location)
	}
	w.WriteHeader(e.Code)
	w.Write([]byte(e.Message + "\n"))
}

func (e *RESTError) Error() string {
	return e.Message
}

type REST struct {
	Init   func(*http.Request) (interface{}, error)
	Delete RESTMethod
	Get    RESTMethod
	Insert RESTMethod
	List   RESTMethod
	Update RESTMethod
}

type RESTMethod func(interface{}, http.ResponseWriter, *http.Request) error

func (s *REST) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var client interface{}
	var err error
	comps := strings.Split(r.URL.Path, "/")
	key := comps[len(comps)-1]
	if s.Init != nil {
		client, err = s.Init(r)
		if err != nil {
			goto Err
		}
	}
	r.URL.Path = key
	switch r.Method {
	case "GET", "HEAD":
		if key == "" {
			err = try(s.List)(client, w, r)
		} else {
			err = try(s.Get)(client, w, r)
		}
	case "POST":
		err = try(s.Insert)(client, w, r)
	case "DELETE":
		err = try(s.Delete)(client, w, r)
	case "PUT":
		err = try(s.Update)(client, w, r)
	default:
		err = RESTMethodNotAllowed
	}
Err:
	switch err := err.(type) {
	case nil:
		// do nothing on success
	case *RESTError:
		err.WriteTo(w)
	default:
		http.Error(w, err.Error()+"\n", http.StatusInternalServerError)
	}
}

func try(m RESTMethod) RESTMethod {
	if m == nil {
		return methodNotAllowed
	}
	return m
}

func methodNotAllowed(_ interface{}, w http.ResponseWriter, r *http.Request) error {
	return RESTMethodNotAllowed
}
