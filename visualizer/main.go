package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"github.com/mesmerai/news-aggregator/visualizer/data"
)

// DB Conn Vars
var environment = getEnv()

var db_port int = 5432
var db_name string = "news"
var db_user string = "news_db_user"
var dbconn_max_retries = 10
var web_user = "carmelo"

// ** Secrets from ENV **
var db_host = environment["db_host"]
var db_password = environment["db_password"]

// Create the JWT Key from  our secret
var jwtKey = []byte(environment["jwt_key"])

// user auth
var web_password = environment["user_auth"]

//DB
var myDB *data.DBClient

// should get those creds from the login form via POST
type Credentials struct {
	Username string
	Password string
}

// this struct will be encoded to a JWT
// We add jwt.StandardClaims to provide fields like expiry time
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Loading HTML Template from file. Will will panic if error is not-nil.
var tmpl = template.Must(template.ParseFiles("./index.html"))

type Data struct {
	Query           string
	NextPage        int
	TotalPages      int
	Results         *data.Results
	Favourites      *data.FavouriteDomains
	NotFavourites   *data.NotFavouriteDomains
	ArticlesPerFeed []data.ArticlePerFeed
	LoggedUser      string
}

var pageData Data

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

	checkToken(w, r)

	// some vars declared
	var err error
	var favResults *data.FavouriteDomains
	var notFavResults *data.NotFavouriteDomains
	var favCount, notFavCount int

	var articlesPerFeed []data.ArticlePerFeed

	// ** retrieve Feeds to populate the menu on the left side **

	favCount = myDB.CountFavouriteDomains()
	favResults = myDB.GetFavouriteDomains()
	favResults.Count = favCount

	notFavCount = myDB.CountNotFavouriteDomains()
	notFavResults = myDB.GetNotFavouriteDomains()
	notFavResults.Count = notFavCount

	// ** articlesPerFeed for the menu on the right **

	articlesPerFeed = myDB.CountArticlesGroupByFavourites() // should return a []ArticlePerFeed having .FeedName and .ArticlesCount

	thisData := &pageData
	*thisData = Data{
		Query:           pageData.Query,
		NextPage:        pageData.NextPage,
		TotalPages:      pageData.TotalPages,
		Results:         pageData.Results,
		Favourites:      favResults,
		NotFavourites:   notFavResults,
		ArticlesPerFeed: articlesPerFeed,
		LoggedUser:      pageData.LoggedUser,
	}

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

}

func authHandler(w http.ResponseWriter, r *http.Request) {
	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
	if err := r.ParseForm(); err != nil {
		log.Fatal("Error Parsing the Form: ", err)
		return
	}

	// Debug
	//log.Printf("Receiving Post Data! r.PostFrom = %v\n", r.PostForm)
	thisUser := r.FormValue("username")
	thisPasswd := r.FormValue("password")
	//log.Println("Userame = ", thisUser)
	//log.Println("Password = ", thisPasswd)

	if web_password == "" {
		log.Fatal("Password cannot be blank.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if thisUser != web_user || thisPasswd != web_password {
		log.Println("Wrong Username or Password.")
		//w.WriteHeader(http.StatusUnauthorized)
		log.Println("Redirecting to Login page.")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 5 minutes expiration time for our token
	expirationTime := time.Now().Add(5 * time.Minute)

	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		Username: thisUser,
		StandardClaims: jwt.StandardClaims{
			// in JWT expire time is expressed in Unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// declare the token with Signign algorithm and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// create the Token String
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// raise an Internal Server Error if there's any error creating the JWT
		w.WriteHeader(http.StatusInternalServerError)
		//return
	}

	// finally set the client cookie with the token and same expiration time
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	// set the username in the pageData for the template logic
	thisData := &pageData
	*thisData = Data{
		Query:           pageData.Query,
		NextPage:        pageData.NextPage,
		TotalPages:      thisData.TotalPages,
		Results:         pageData.Results,
		Favourites:      pageData.Favourites,
		NotFavourites:   pageData.NotFavourites,
		ArticlesPerFeed: pageData.ArticlesPerFeed,
		LoggedUser:      thisUser,
	}

	log.Println("Token set.")
	log.Println("Redirecting to main page.")
	http.Redirect(w, r, "/", http.StatusFound)

}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	var err error
	// define empty intermediate buffer
	buffer := &bytes.Buffer{}

	// write to intermediate buffer to check errors
	// NOTE that I'm passing nil instead of pageData
	err = tmpl.Execute(buffer, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Then the buffer is written to the ResponseWriter
	// func (r *Reader) WriteTo(w io.Writer) (n int64, err error)
	buffer.WriteTo(w)

}

func addFeedsHandler(w http.ResponseWriter, r *http.Request) {

	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	// require valid token
	checkToken(w, r)

	// some vars declared
	var err error

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

	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	// require valid token
	checkToken(w, r)

	// some vars declared
	var err error

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

	// log the request
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	// require valid token
	checkToken(w, r)

	// some vars declared
	var err error
	var results *data.Results
	var count int
	//results := &data.Results{}

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
		log.Fatal("Not a valid search parameter")
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
		Query:           searchQuery,
		NextPage:        nextPage,
		TotalPages:      tot,
		Results:         results,
		Favourites:      pageData.Favourites,
		NotFavourites:   pageData.NotFavourites,
		ArticlesPerFeed: pageData.ArticlesPerFeed,
		LoggedUser:      pageData.LoggedUser,
	}

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

func checkToken(w http.ResponseWriter, r *http.Request) {

	// ** Authentication Check **

	// get the token from the cookie, that comes at every request
	c, cookieErr := r.Cookie("token")
	if cookieErr != nil {
		if cookieErr == http.ErrNoCookie {
			// Not Authorized
			//w.WriteHeader(http.StatusUnauthorized)

			// redirect to login page
			log.Printf("Unauthorized Access => %s", cookieErr)
			log.Println("Redirecting to Login page.")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// For any other err, it's Bad Request
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Bad Request: ", cookieErr)
		return
	}

	// get the token from the Cookie
	tokenStr := c.Value

	// Initialize an instance of Claims
	claims := &Claims{}

	// Parse the JWT string and store it in claims
	// We pass the key as well
	// this method will return error if token is expired or key doesn't match
	tkn, tknErr := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if tknErr != nil {
		if tknErr == jwt.ErrSignatureInvalid {
			//w.WriteHeader(http.StatusUnauthorized)
			// redirect to login page
			log.Printf("Token Signature Invalid => %s", tknErr)
			log.Println("Redirecting to Login page.")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Bad Request: ", cookieErr)
		return
	}

	if !tkn.Valid {
		//w.WriteHeader(http.StatusUnauthorized)
		log.Println("Token isn't valid.")
		log.Println("Redirecting to Login page.")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// >> At this stage can print a Welcome message to the user <<

	// ** END Authentication Check **
}

func getEnv() map[string]string {

	envMap := map[string]string{}

	db_host := os.Getenv("DB_HOST")
	if db_host == "" {
		log.Fatal("DB_HOST is not set in ENV.")
	}
	db_password := os.Getenv("DB_PASSWORD")
	if db_password == "" {
		log.Fatal("Password for the DB is not set in ENV.")
	}
	jwt_key := os.Getenv("JWT_KEY")
	if db_password == "" {
		log.Fatal("JWT_KEY is not set in ENV.")
	}
	user_auth := os.Getenv("USER_AUTH")
	if db_password == "" {
		log.Fatal("USER_AUTH is not set in ENV.")
	}

	envMap["db_host"] = db_host
	envMap["db_password"] = db_password
	envMap["jwt_key"] = jwt_key
	envMap["user_auth"] = user_auth

	return envMap
}

func main() {

	/* ** DB Conn ** */
	myDB = data.NewDBClient(db_host, db_port, db_name, db_user, db_password, dbconn_max_retries)

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
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/auth", authHandler)
	// static files Handle
	// use Handle because the http.FileServer() method returns an http.Handler type instead of an HandlerFunc
	// we Strip the prefix to cut the '/assets/' part and forward the modified request to the handler
	//   returned by http.FileServer() so it will see the requested resource as style.css
	//   and look at it in the ./assets folder then
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// handler for /search
	mux.HandleFunc("/search", searchHandler)

	mux.HandleFunc("/addFeeds", addFeedsHandler)
	mux.HandleFunc("/saveFeeds", saveFeedsHandler)

	// ListenAndServe starts an HTTP server with a given address and handler.
	// -- http://localhost:8080
	http.ListenAndServe(":8080", mux)

	log.Println("Server Listening.")
}
