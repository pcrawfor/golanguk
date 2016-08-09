package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/gorilla/websocket"
	"github.com/pcrawfor/golanguk/samples/app/user"
)

const giphyKey = "dc6zaTOxFJmzC"

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
	wd, _ := os.Getwd()
	dir := wd
	templates = template.Must(template.ParseGlob(dir + "/templates/*.html"))
	templates.ExecuteTemplate(w, "home", "ws://"+r.Host+"/echo")
	//homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
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

	log.Fatal(http.ListenAndServe(":3000", nil))
}

// var homeTemplate = template.Must(template.New("").Parse(`
// <!DOCTYPE html>
// <head>
// <meta charset="utf-8">
// <script>
// window.addEventListener("load", function(evt) {
//     var output = document.getElementById("output");
//     var input = document.getElementById("input");
//     var ws;
//     var print = function(message) {
//         var d = document.createElement("div");
//         d.innerHTML = message;
//         output.appendChild(d);
//     };
//     document.getElementById("open").onclick = function(evt) {
//         if (ws) {
//             return false;
//         }
//         ws = new WebSocket("{{.}}");
//         ws.onopen = function(evt) {
//             print("OPEN");
//         }
//         ws.onclose = function(evt) {
//             print("CLOSE");
//             ws = null;
//         }
//         ws.onmessage = function(evt) {
//             print("RESPONSE: " + evt.data);
//         }
//         ws.onerror = function(evt) {
//             print("ERROR: " + evt.data);
//         }
//         return false;
//     };
//     document.getElementById("send").onclick = function(evt) {
//         if (!ws) {
//             return false;
//         }
//         print("SEND: " + input.value);
//         ws.send(input.value);
//         return false;
//     };
//     document.getElementById("close").onclick = function(evt) {
//         if (!ws) {
//             return false;
//         }
//         ws.close();
//         return false;
//     };
// });
// </script>
// </head>
// <body>
// <table>
// <tr><td valign="top" width="50%">
// <p>Click "Open" to create a connection to the server,
// "Send" to send a message to the server and "Close" to close the connection.
// You can change the message and send multiple times.
// <p>
// <form>
// <button id="open">Open</button>
// <button id="close">Close</button>
// <p><input id="input" type="text" value="Hello world!">
// <button id="send">Send</button>
// </form>
// </td><td valign="top" width="50%">
// <div id="output"></div>
// </td></tr></table>
// </body>
// </html>
// `))
