package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"context"

	"github.com/gorilla/sessions"
	"github.com/pcrawfor/golanguk/samples/app/auth"
	"github.com/pcrawfor/golanguk/samples/app/lookup"
	"github.com/pcrawfor/golanguk/samples/app/session"
)

var templates *template.Template
var store *sessions.CookieStore

// this is a public API key available from the giphy github page
const giphyKey = "dc6zaTOxFJmzC"

func login(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login", nil)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session.Delete(w, r, store)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if auth.Authenticate(email, password) {
		session.Save(email, w, r, store)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func home(w http.ResponseWriter, r *http.Request) {
	s, err := session.FromRequest(r, store)
	user, ok := session.Email(s)
	if err != nil || !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	params := map[string]interface{}{
		"user": user,
	}

	templates.ExecuteTemplate(w, "search", params)
}

func search(w http.ResponseWriter, r *http.Request) {
	reloadTemplates()

	s, err := session.FromRequest(r, store)
	user, ok := session.Email(s)
	if err != nil || !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	qry := r.FormValue("input")

	// create a context with a hard deadline for returning something
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	type resultAndError struct {
		results []string
		err     error
	}

	// ask duckduckgo for an answer
	answerChan := make(chan resultAndError)
	go func() {
		value, err := lookup.DuckduckQuery(ctx, qry)
		answerChan <- resultAndError{value, err}
	}()

	// ask giphy for a gif
	sessionCtx := session.NewContext(ctx, s)
	gifChan := make(chan resultAndError)
	go func() {
		terms := strings.Split(qry, " ")
		url, err := lookup.GifForTerms(sessionCtx, terms, giphyKey)
		gifChan <- resultAndError{[]string{url}, err}
	}()

	var results []string
	var gif string

	func() {
		for {
			select {
			case r := <-answerChan:
				results = r.results
				if r.err != nil || len(results) < 1 {
					results = []string{"Whoops we couldn't find anything!"}
				}
				log.Println("CANCEL GIF REQUEST")
				cancel()
				return
			case r := <-gifChan:
				if r.err != nil {
					continue
				}
				gif = r.results[0]
			case <-ctx.Done():
				results = []string{"Whoops we ran out of time!"}
				return
			}
		}
	}()

	params := map[string]interface{}{
		"results":  results,
		"question": qry,
		"gif":      gif,
		"user":     user,
	}

	templates.ExecuteTemplate(w, "results", params)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	store = sessions.NewCookieStore([]byte("somemagichash-askjhdsqhwesdfjh13u234skdfds"))

	reloadTemplates()

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", home)
	http.HandleFunc("/search", search)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/authenticate", authenticate)

	log.Fatal(http.ListenAndServe(":3000", nil))
}

func reloadTemplates() {
	wd, _ := os.Getwd()
	dir := wd
	templates = template.Must(template.ParseGlob(dir + "/templates/*.html"))
}
