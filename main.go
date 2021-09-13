package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
)

// Loading HTML Template from file. Will will panic if error is not-nil.
var tmpl = template.Must(template.ParseFiles("./index.html"))

/* * Handler function *
- define the IndexHandler function having signature => func(w http.ResponseWriter, r *http.Request)
- The w parameter is the structure we use to send responses to an HTTP request.
- It implements a Write() method which accepts a slice of bytes and writes the data
to the connection as part of an HTTP response.
- the r parameter represents the HTTP request received from the client.
*/
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	//show the HTML template
	tmpl.Execute(w, nil)
	//w.Write([]byte("<h1>Hello World!</h1>\n"))
	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Package url parses URLs and implements query escaping ==> http://localhost:8080/search?q=ciccio
	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := u.Query()
	searchQuery := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	fmt.Println("Search Query: ", searchQuery)
	fmt.Println("Page: ", page)

}

func main() {

	// to handle static files (like our assets/style.css) we need to:
	// - instantiate a FileServer object with the folder of the static files
	fs := http.FileServer(http.Dir("./assets"))
	// - then add a Handle to the Router for the '/assets/' prefix (see below)

	// create a HTTP request Multiplexer
	mux := http.NewServeMux()

	// add Handles, basically matches Requests and call the respective Handle
	mux.HandleFunc("/", IndexHandler)
	// static files Handle
	// use Handle because the http.FileServer() method returns an http.Handler type instead of an HandlerFunc
	// we Strip the prefix to cut the '/assets/' part and forward the modified request to the handler
	//   returned by http.FileServer() so it will see the requested resource as style.css
	//   and look at it in the ./assets folder then
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// handler for /search
	mux.HandleFunc("/search", SearchHandler)

	// ListenAndServe starts an HTTP server with a given address and handler.
	// -- http://localhost:8080
	http.ListenAndServe(":8080", mux)

}

func getApiKeyFromEnv() string {
	news_api_key := os.Getenv("NEWS_API_KEY")
	if news_api_key == "" {
		log.Fatal("News Api Key is not set in ENV.")
	}

	return news_api_key
}
