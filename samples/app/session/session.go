package session

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

const userSessionKey string = "user"
const storeKey string = "session-user"

type key int

const userCtxKey key = 0

func getSession(req *http.Request, store *sessions.CookieStore) (*sessions.Session, error) {
	if store != nil {
		return store.Get(req, storeKey)
	}
	return nil, nil
}

func Save(email string, rw http.ResponseWriter, req *http.Request, store *sessions.CookieStore) error {
	log.Println("SAVE session")
	session, err := getSession(req, store)
	if err != nil {
		log.Println("SAVE session:", err)
		return err
	}

	session.Values[userSessionKey] = email
	return session.Save(req, rw)
}

func Delete(rw http.ResponseWriter, req *http.Request, store *sessions.CookieStore) error {
	session, err := getSession(req, store)
	if err != nil {
		return err
	}
	delete(session.Values, userSessionKey)
	return session.Save(req, rw)
}

// FromRequest extracts the session from req, if present.
func FromRequest(req *http.Request, store *sessions.CookieStore) (string, error) {
	log.Println("FromRequest")
	if store == nil {
		log.Println("FromRequest store invalid")
		return "", errors.New("Cookie store is nil")
	}

	s, e := getSession(req, store)
	if e != nil {
		log.Println("FromRequest:", e)
		return "", e
	}
	email, ok := s.Values[userSessionKey].(string)
	if ok {
		log.Println("Email:", email)
		return email, nil
	}

	log.Println("FromRequest NOT FOUND")
	return "", errors.New("No Email found")
}

// NewContext returns a new Context carrying user email
func NewContext(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userCtxKey, email)
}

// FromContext extracts the user email address from ctx, if present.
func FromContext(ctx context.Context) (string, bool) {
	// ctx.Value returns nil if ctx has no value for the key
	email, ok := ctx.Value(userCtxKey).(string)
	return email, ok
}
