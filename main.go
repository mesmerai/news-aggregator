package main

import (
	"bytes"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/mesmerai/news-aggregator/news"
)

// Only needed with * Approach 1 * to make newsapi accessible to SearchHandler
//var newsapi *news.Client

// Loading HTML Template from file. Will will panic if error is not-nil.
var tmpl = template.Must(template.ParseFiles("./index.html"))

// Search struct for search queries. Populated and used in the html template as data object in the searchHandler
type Search struct {
	Query      string
	NextPage   int
	TotalPages int
	Results    *news.Results
}

// determine if it's LastPage to set the 'Next' button
func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

// get CurrentPage to set the 'Previous' button
// CurrentPage is always NextPage - 1, except when we have only one page
func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}
	return s.NextPage - 1
}

// PreviousPage
func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}

/* * Handler function *
- define the IndexHandler function having signature => func(w http.ResponseWriter, r *http.Request)
- The w parameter is the structure we use to send responses to an HTTP request.
- It implements a Write() method which accepts a slice of bytes and writes the data
to the connection as part of an HTTP response.
- the r parameter represents the HTTP request received from the client.
*/
func indexHandler(w http.ResponseWriter, r *http.Request) {
	//show the HTML template
	tmpl.Execute(w, nil)
	//w.Write([]byte("<h1>Hello World!</h1>\n"))
	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
}

/* * Approach 2 *
   * Closure / Functions declared inside of functions are special; they are closures. *

	1. newsapi defined and assigned inside main
	2. use a Closure to pass the newsapi inside searchHandler => mux.HandleFunc("/search", searchHandler(newsapi))
	3. searchHandler function must be adjusted to include the Closure as per below

	searchHandler receive a Pointer to news.Client and return an anonymous function that satisfies the HandlerFunc

	NOTE: This is potentially a better solution since:
	- it makes testing much easier
	- limits the function's scope
*/
func searchHandler(newsapi *news.Client, searchType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// some vars declared
		var err error
		country := ""
		results := &news.Results{}

		// log the request
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

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

		//fmt.Println("Search Query: ", searchQuery)
		//fmt.Println("Page: ", page)

		if searchType == "searchGlobal" {
			// the Global search returns error in the APi if "q" is not set as the scope of your search is too broad
			if searchQuery == "" {
				log.Fatal("Scope of the search is too broad, must specify a keyword.")
			}
			results, err = newsapi.FetchNews(searchQuery, page, "")
		}
		if searchType == "searchByCountry" {
			// the searchByCountry returns error in the API if "country" is not set
			country = params.Get("country")
			if country == "" {
				log.Fatal("Country must specified in SearchByCountry.")
			}
			results, err = newsapi.FetchNews(searchQuery, page, country)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Printing the Results from the Search, the whole struct
		// %v	the value in a default format
		// when printing structs, the plus flag (%+v) adds field names
		//fmt.Printf("%+v", results)

		// we convert page into int first
		nextPage, err := strconv.Atoi(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// We save our results into the Search struct defined above
		// so that we can use it for Pagination
		search := &Search{
			Query:      searchQuery,
			NextPage:   nextPage,
			TotalPages: int(math.Ceil(float64(results.TotalResults) / float64(newsapi.PageSize))), //rounding the result up to the nearest integer, used later for pagination
			Results:    results,
		}

		// this block is to increment NextPage
		if !search.IsLastPage() {
			search.NextPage++
		}

		// Intermediatie empty byte buffer where the Template is execute first to check errors
		// a Buffer is a struct and needs no initialization - ref. https://pkg.go.dev/bytes#Buffer
		buffer := &bytes.Buffer{}

		// We write the template to to the empty byte.Buffer passing the the 'search' data object
		// func (t *Template) Execute(wr io.Writer, data interface{}) error
		// Execute applies a parsed template to the specified data object, writing the output to wr.
		// If an error occurs executing the template or writing its output, execution stops, but partial results may already have been written to the output writer.
		err = tmpl.Execute(buffer, search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Then the buffer is written to the ResponseWriter
		// func (r *Reader) WriteTo(w io.Writer) (n int64, err error)
		buffer.WriteTo(w)

	}
}

/*
func searchByCountryHandler(newsapi *news.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// log the request
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

		// pare the URL first http://localhost:8080/searchByCountry?country=Italy&q=devops
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		params := u.Query()
		country := params.Get("country")
		searchQuery := params.Get("q")
		page := params.Get("page")
		if page == "" {
			page = "1"
		}

		results, err := newsapi.FetchNewsByCountry(searchQuery, page, country)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextPage, err := strconv.Atoi(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		search := &Search{
			Query:      searchQuery,
			NextPage:   nextPage,
			TotalPages: int(math.Ceil(float64(results.TotalResults) / float64(newsapi.PageSize))), //rounding the result up to the nearest integer, used later for pagination
			Results:    results,
		}

		if !search.IsLastPage() {
			search.NextPage++
		}

		buffer := &bytes.Buffer{}

		err = tmpl.Execute(buffer, search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buffer.WriteTo(w)
	}

}
*/

func getApiKeyFromEnv() string {
	news_api_key := os.Getenv("NEWS_API_KEY")
	if news_api_key == "" {
		log.Fatal("News Api Key is not set in ENV.")
	}

	return news_api_key
}

func main() {
	news_api_key := getApiKeyFromEnv()
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 20)

	// to handle static files (like our assets/style.css) we need to:
	// - instantiate a FileServer object with the folder of the static files
	fs := http.FileServer(http.Dir("./assets"))
	// - then add a Handle to the Router for the '/assets/' prefix (see below)

	// create a HTTP request Multiplexer
	mux := http.NewServeMux()

	// add Handles, basically matches Requests and call the respective Handle
	mux.HandleFunc("/", indexHandler)
	// static files Handle
	// use Handle because the http.FileServer() method returns an http.Handler type instead of an HandlerFunc
	// we Strip the prefix to cut the '/assets/' part and forward the modified request to the handler
	//   returned by http.FileServer() so it will see the requested resource as style.css
	//   and look at it in the ./assets folder then
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// handler for /search
	// ** Closure over newsapi parameter
	mux.HandleFunc("/search", searchHandler(newsapi, "searchGlobal"))
	mux.HandleFunc("/searchByCountry", searchHandler(newsapi, "searchByCountry"))

	// searchNews by Country
	//mux.HandleFunc("/searchByCountry", searchByCountryHandler(newsapi))

	// ListenAndServe starts an HTTP server with a given address and handler.
	// -- http://localhost:8080
	http.ListenAndServe(":8080", mux)
}
