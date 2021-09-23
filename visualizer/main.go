package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/mesmerai/news-aggregator/visualizer/data"
)

// DB Conn Vars
var environment = getEnv()
var db_host string = "localhost"
var db_port int = 5432
var db_name string = "news"
var db_user string = "news_db_user"

// Secrets from ENV
var db_password = environment["db_password"]

//DB
var myDB *data.DBClient

// Loading HTML Template from file. Will will panic if error is not-nil.
var tmpl = template.Must(template.ParseFiles("./index.html"))

type Data struct {
	Query         string
	NextPage      int
	TotalPages    int
	Results       *data.Results
	Favourites    *data.FavouriteDomains
	NotFavourites *data.NotFavouriteDomains
}

var pageData Data

/*
type Feeds struct {
	Favourites    *data.FavouriteDomains
	NotFavourites *data.NotFavouriteDomains
}

// Search struct for search queries. Populated and used in the html template as data object in the searchHandler
type Search struct {
	Query      string
	NextPage   int
	TotalPages int
	Results    *data.Results
}
*/

// determine if it's LastPage to set the 'Next' button
func (s *Data) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

// get CurrentPage to set the 'Previous' button
// CurrentPage is always NextPage - 1, except when we have only one page
func (s *Data) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}
	return s.NextPage - 1
}

// PreviousPage
func (s *Data) PreviousPage() int {
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

	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	// some vars declared
	var err error
	var favResults *data.FavouriteDomains
	var notFavResults *data.NotFavouriteDomains
	var favCount, notFavCount int

	// ** retrieve Feeds to populate the menu on the left side */

	favCount = myDB.CountFavouriteDomains()
	favResults = myDB.GetFavouriteDomains()
	favResults.Count = favCount

	notFavCount = myDB.CountNotFavouriteDomains()
	notFavResults = myDB.GetNotFavouriteDomains()
	notFavResults.Count = notFavCount

	thisData := &pageData
	*thisData = Data{
		Query:         pageData.Query,
		NextPage:      pageData.NextPage,
		TotalPages:    pageData.TotalPages,
		Results:       pageData.Results,
		Favourites:    favResults,
		NotFavourites: notFavResults,
	}

	/*
		data := &Data{
			Query:         &Data.Query,
			NextPage:      NextPage,
			TotalPages:    TotalPages,
			Results:       Results,
			Favourites:    favResults,
			NotFavourites: notFavResults,
		}
	*/
	// define empty intermediate buffer
	buffer := &bytes.Buffer{}

	// write to intermediate buffer to check errors
	err = tmpl.Execute(buffer, thisData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Then the buffer is written to the ResponseWriter
	// func (r *Reader) WriteTo(w io.Writer) (n int64, err error)
	buffer.WriteTo(w)

	//show the HTML template
	//tmpl.Execute(w, nil)

	//w.Write([]byte("<h1>Hello World!</h1>\n"))

}

func addFeedsHandler(w http.ResponseWriter, r *http.Request) {

	// some vars declared
	var err error

	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	// Package url parses URLs and implements query escaping ==> http://localhost:8080/search?q=ciccio
	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := u.Query()
	feeds := params["afeed"]

	//fmt.Println(feeds)

	myDB.SetFavourites(feeds)

	// redirect to root
	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

func saveFeedsHandler(w http.ResponseWriter, r *http.Request) {

	// some vars declared
	var err error

	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	// Package url parses URLs and implements query escaping ==> http://localhost:8080/saveFeeds?feed=adnkronos.com&feed=ansa.it
	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := u.Query()
	feeds := params["sfeed"]
	//dList := make([]string, 0)

	//fmt.Println(feeds)

	myDB.ResetFavourites()
	myDB.SetFavourites(feeds)

	// redirect to root
	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	// some vars declared
	var err error
	var results *data.Results
	var count int
	//results := &data.Results{}

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
	limit := params.Get("limit")
	country := params.Get("country")

	// set defaults if param is missing
	if page == "" {
		page = "1"
	}

	if limit == "" {
		limit = "100"
	}

	if country == "" {
		country = "Global"
	}

	// convert to int some params
	limitToInt, err := strconv.Atoi(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageToInt, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//limit := 100
	offset := (pageToInt * limitToInt) - limitToInt

	// call Global Search
	switch {
	case country == "Global":
		if searchQuery == "" {
			count = myDB.CountArticles("")
			results = myDB.GetArticles(limitToInt, offset, "")
		} else {
			count = myDB.CountArticles(searchQuery)
			results = myDB.GetArticles(limitToInt, offset, searchQuery)
		}
		results.TotalResults = count
	case country == "Australia":
		if searchQuery == "" {
			count = myDB.CountArticlesByCountry(country, "")
			results = myDB.GetArticlesByCountry(limitToInt, offset, country, "")
		} else {
			count = myDB.CountArticlesByCountry(country, searchQuery)
			results = myDB.GetArticlesByCountry(limitToInt, offset, country, searchQuery)
		}
		results.TotalResults = count
	case country == "Italy":
		if searchQuery == "" {
			count = myDB.CountArticlesByCountry(country, "")
			results = myDB.GetArticlesByCountry(limitToInt, offset, country, "")
		} else {
			count = myDB.CountArticlesByCountry(country, searchQuery)
			results = myDB.GetArticlesByCountry(limitToInt, offset, country, searchQuery)
		}
		results.TotalResults = count
	default:

	}
	// we convert page into int first
	nextPage, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// calculate total pages
	var tot int
	mod := results.TotalResults % limitToInt
	if mod == 0 {
		tot = results.TotalResults / limitToInt
	} else {
		tot = (results.TotalResults / limitToInt) + 1
	}

	// We save our results into the Search struct defined above
	// so that we can use it for Pagination

	thisData := &pageData
	*thisData = Data{
		Query:         searchQuery,
		NextPage:      nextPage,
		TotalPages:    tot,
		Results:       results,
		Favourites:    pageData.Favourites,
		NotFavourites: pageData.NotFavourites,
	}

	/*
		data := &Data{
			Query:      searchQuery,
			NextPage:   nextPage,
			TotalPages: tot,
			Results:    results,
		}
	*/

	// this block is to increment NextPage
	if !thisData.IsLastPage() {
		thisData.NextPage++
	}

	// Intermediatie empty byte buffer where the Template is execute first to check errors
	// a Buffer is a struct and needs no initialization - ref. https://pkg.go.dev/bytes#Buffer
	buffer := &bytes.Buffer{}

	// We write the template to to the empty byte.Buffer passing the the 'search' data object
	// func (t *Template) Execute(wr io.Writer, data interface{}) error
	// Execute applies a parsed template to the specified data object, writing the output to wr.
	// If an error occurs executing the template or writing its output, execution stops, but partial results may already have been written to the output writer.
	err = tmpl.Execute(buffer, thisData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Then the buffer is written to the ResponseWriter
	// func (r *Reader) WriteTo(w io.Writer) (n int64, err error)
	buffer.WriteTo(w)

}

func main() {

	/* ** DB Conn ** */
	myDB = data.NewDBClient(db_host, db_port, db_name, db_user, db_password)

	// myDB = *DBClient(db_conn)
	defer myDB.Database.Close()

	log.Println("Closing DB resources.")

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
	mux.HandleFunc("/search", searchHandler)

	mux.HandleFunc("/addFeeds", addFeedsHandler)
	mux.HandleFunc("/saveFeeds", saveFeedsHandler)

	// ListenAndServe starts an HTTP server with a given address and handler.
	// -- http://localhost:8080
	http.ListenAndServe(":8080", mux)
}

func getEnv() map[string]string {

	envMap := map[string]string{}

	news_api_key := os.Getenv("NEWS_API_KEY")
	if news_api_key == "" {
		log.Fatal("News Api Key is not set in ENV.")
	}
	db_password := os.Getenv("DB_PASSWORD")
	if db_password == "" {
		log.Fatal("Password for the DB is not set in ENV.")
	}

	envMap["news_api_key"] = news_api_key
	envMap["db_password"] = db_password

	return envMap
}
