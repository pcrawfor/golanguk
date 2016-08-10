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

//var addr = flag.String("addr", "localhost:3000", "http service address")
var upgrader = websocket.Upgrader{} // use default options

var templates *template.Template

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// TODO: review the structure
	// TODO: extract request information and set it as context value(s)
	// TODO: think of an external service integration that makes sense to demo to show deadline/timeout?

	us := user.NewSession()
	ctx, cancel := context.WithCancel(context.Background())
	us.Start(ctx, c)
	defer cancel()
}

func home(w http.ResponseWriter, r *http.Request) {
	reloadTemplates()
	templates.ExecuteTemplate(w, "home", "ws://"+r.Host+"/echo")
}

func search(w http.ResponseWriter, r *http.Request) {
	reloadTemplates()
	templates.ExecuteTemplate(w, "search", nil)
}

func ask(w http.ResponseWriter, r *http.Request) {
	qry := r.FormValue("input")

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	defer cancel()

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

	gifCtx, gifCancel := context.WithCancel(ctx)

	// ask giphy for a gif
	gifChan := make(chan resultAndError)
	go func() {
		g := lookup.NewGiphy(user.GiphyKey)
		terms := strings.Split(qry, " ")
		url, err := g.GifForTerms(gifCtx, terms)
		gifChan <- resultAndError{url, err}
	}()

	var result, gif string

	func() {
		for {
			select {
			case r := <-answerChan:
				result = r.result
				if r.err != nil {
					result = "Whoops we couldn't find anything!"
				}
				log.Println("CANCEL GIF REQUEST")
				gifCancel()
				return
			case r := <-gifChan:
				if r.err != nil {
					continue
				}
				gif = r.result
			case <-ctx.Done():
				result = "Whoops we couldn't find anything!"
				return
			}
		}
	}()

	params := map[string]interface{}{
		"result":   result,
		"question": qry,
		"gif":      gif,
	}

	log.Println("Result:", result)
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
	//    dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	templates = template.Must(template.ParseGlob(dir + "/templates/*.html"))

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	http.HandleFunc("/search", search)
	http.HandleFunc("/ask", ask)

	log.Fatal(http.ListenAndServe(":3000", nil))
}
