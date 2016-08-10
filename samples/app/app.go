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

	"github.com/gorilla/websocket"
	"github.com/pcrawfor/golanguk/samples/app/lookup"
	"github.com/pcrawfor/golanguk/samples/app/user"
)

var upgrader = websocket.Upgrader{} // use default options

var templates *template.Template

func websock(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	us := user.NewSession()
	ctx, cancel := context.WithCancel(context.Background())
	us.Start(ctx, c)
	defer cancel()
}

func home(w http.ResponseWriter, r *http.Request) {
	reloadTemplates()
	templates.ExecuteTemplate(w, "home", "ws://"+r.Host+"/websock")
}

func search(w http.ResponseWriter, r *http.Request) {
	reloadTemplates()
	templates.ExecuteTemplate(w, "search", nil)
}

func ask(w http.ResponseWriter, r *http.Request) {
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
		g := lookup.NewGiphy(user.GiphyKey)
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
	}

	templates.ExecuteTemplate(w, "results", params)
}

func reloadTemplates() {
	wd, _ := os.Getwd()
	dir := wd
	templates = template.Must(template.ParseGlob(dir + "/templates/*.html"))
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	wd, _ := os.Getwd()
	dir := wd

	templates = template.Must(template.ParseGlob(dir + "/templates/*.html"))

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", search)
	http.HandleFunc("/websock", websock)
	http.HandleFunc("/search", search)
	http.HandleFunc("/ask", ask)

	log.Fatal(http.ListenAndServe(":3000", nil))
}
