package session

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

const userSessionKey string = "user"
const storeKey string = "session-user"

// Save stores the email value on the session
func Save(email string, rw http.ResponseWriter, req *http.Request, store *sessions.CookieStore) error {
	session, err := FromRequest(req, store)
	if err != nil {
		return err
	}

	session.Values[userSessionKey] = email
	return session.Save(req, rw)
}

// Delete removes the email from the session
func Delete(rw http.ResponseWriter, req *http.Request, store *sessions.CookieStore) error {
	session, err := FromRequest(req, store)
	if err != nil {
		return err
	}
	delete(session.Values, userSessionKey)
	return session.Save(req, rw)
}

// Email extracts and returns the email from the session along with a boolean indicating whether there is an email value stored
func Email(s *sessions.Session) (string, bool) {
	email, ok := s.Values[userSessionKey].(string)
	if ok {
		return email, true
	}

	return "", false
}

type key int

const sessionCtxKey key = 0

// FromRequest extracts the user email from req, if present.
func FromRequest(req *http.Request, store *sessions.CookieStore) (*sessions.Session, error) {
	if store == nil {
		return nil, errors.New("Cookie store is nil")
	}

	return store.Get(req, storeKey)
}

// NewContext returns a new Context carrying session
func NewContext(ctx context.Context, s *sessions.Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, s)
}

// FromContext extracts the session from ctx, if present.
func FromContext(ctx context.Context) (*sessions.Session, bool) {
	// ctx.Value returns nil if ctx has no value for the key
	s, ok := ctx.Value(sessionCtxKey).(*sessions.Session)
	return s, ok
}
