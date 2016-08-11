package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/gorilla/sessions"
	"github.com/pcrawfor/golanguk/samples/app/auth"
	"github.com/pcrawfor/golanguk/samples/app/lookup"
	"github.com/pcrawfor/golanguk/samples/app/session"
)

var templates *template.Template
var store *sessions.CookieStore

const GiphyKey = "dc6zaTOxFJmzC"

func login(w http.ResponseWriter, r *http.Request) {
	reloadTemplates()
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

func search(w http.ResponseWriter, r *http.Request) {
	user, err := session.FromRequest(r, store)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	reloadTemplates()

	params := map[string]interface{}{
		"user": user,
	}

	templates.ExecuteTemplate(w, "search", params)
}

func ask(w http.ResponseWriter, r *http.Request) {
	user, err := session.FromRequest(r, store)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	qry := r.FormValue("input")

	// create a context with a hard deadline for returning something
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	type resultAndError struct {
		result string
		err    error
	}

	// ask duckduckgo for an answer
	answerChan := make(chan resultAndError)
	go func() {
		result, err := lookup.DuckduckQuery(ctx, qry)
		answerChan <- resultAndError{result, err}
	}()

	// ask giphy for a gif
	gifChan := make(chan resultAndError)
	go func() {
		g := lookup.NewGiphy(GiphyKey)
		terms := strings.Split(qry, " ")
		url, err := g.GifForTerms(ctx, terms)
		gifChan <- resultAndError{url, err}
	}()

	var result, gif string

	func() {
		for {
			select {
			case r := <-answerChan:
				result = r.result
				if r.err != nil || len(result) < 1 {
					result = "Whoops we couldn't find anything!"
				}
				log.Println("CANCEL GIF REQUEST")
				cancel()
				return
			case r := <-gifChan:
				if r.err != nil {
					continue
				}
				gif = r.result
			case <-ctx.Done():
				result = "Whoops we ran out of time!"
				return
			}
		}
	}()

	params := map[string]interface{}{
		"result":   result,
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

	http.HandleFunc("/", search)
	http.HandleFunc("/search", search)
	http.HandleFunc("/ask", ask)
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
