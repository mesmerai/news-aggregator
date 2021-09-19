package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/mesmerai/news-aggregator/news-collector/news"
)

var environment = getEnv()
var db_host string = "localhost"
var db_port int = 5432
var db_name string = "news"
var db_user string = "news_db_user"

// from Env
var db_password = environment["db_password"]
var news_api_key string = environment["news_api_key"]

func NewDbConnector(db_host string, db_port int, db_name string, db_user string, db_password string) (db *sql.DB) {

	// currently 'sslmode=verify-full' gives error: "sslmode=verify-full"
	connection_string := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", db_host, db_port, db_user, db_password, db_name)

	log.Println("Initiate Connection to DB.")

	// sql.Open() simply validates the arguments provided, doesn't connect yet!
	db_conn, err := sql.Open("postgres", connection_string)
	if err != nil {
		log.Fatal("Error validating DB connection parameters => ", err)
	}

	//defer db_conn.Close()

	// the method Ping() is actually attempting a connection to the database
	err = db_conn.Ping()
	if err != nil {
		log.Fatal("Error connecting to DB => ", err)
	}

	log.Println("Connection to DB successful.")
	return db_conn

}

func InsertSource(db *sql.DB, sourceName string) (sourceID int) {

	log.Printf("Initiate InsertSource for %s", sourceName)

	id := 0
	var selectRows *sql.Rows
	var insertRow *sql.Row
	var selectErr, insertErr error
	sqlSelect := ""
	sqlInsert := ""

	// there's nothing provided in the Sql package to check if Rows has no records inside
	// so I define an empty slice and fill it in the iteration below
	sources := make([]string, 0)

	// Source struct to dealt with INSERT later. Declare and initialize.
	type Source struct {
		id   int
		name string
	}
	var thisSource = Source{}

	log.Println("Checking if the source is the DB already.")

	// check if Source exists
	sqlSelect = "SELECT * FROM sources WHERE name = $1"

	selectRows, selectErr = db.Query(sqlSelect, sourceName)
	if selectErr != nil {
		log.Fatal("Error on SQL SELECT => ", selectErr)
	}

	// rows is a struct - https://cs.opensource.google/go/go/+/refs/tags/go1.17.1:src/database/sql/sql.go;l=2875
	// Rows is the result of a query.
	// Its cursor starts before the first row of the result set. Use Next to advance from row to row.
	for selectRows.Next() {

		// Scan copies the columns in the current row into the values pointed at by dest.
		// The number of values in dest must be the same as the number of columns in Rows.
		err := selectRows.Scan(&thisSource.id, &thisSource.name)
		if err != nil {
			log.Fatal("Error on reading SQL SELECT results => ", err)
		}
		// fill the slice to check if there are rows later
		sources = append(sources, thisSource.name)
		log.Printf("Records found for source: '%s'", sourceName)
		//log.Printf(" - rows.id: %d\n", thisSource.id)
		//log.Printf(" - rows.name: %s\n", thisSource.name)

	}
	log.Println("Closing rows resources.")
	defer selectRows.Close()

	// the 'source' slide is empty if no rows are returned by the SELECT
	if len(sources) == 0 {
		log.Printf("No sources found for '%s'. Proceed with INSERT.", sourceName)

		// INSERT INTO source (name) VALUES ('Ansa');
		sqlInsert = `INSERT INTO sources (name) VALUES ($1) RETURNING id`

		// QueryRow returns a *Row
		insertRow = db.QueryRow(sqlInsert, sourceName)
		insertErr = insertRow.Scan(&id)
		if insertErr != nil {
			log.Fatal("Error on SQL INSERT => ", insertErr)
		}

		log.Printf("Source '%s' stored in the DB.", sourceName)

		// return the id of the source just INSERTed
		return id

	} else {

		log.Println("The source already exists. No INSERT required.")

		// return the id of existing source, so that we can link as foreign key in Article
		return thisSource.id
	}

}

func InsertArticle(db *sql.DB, sourceID int, author string, title string, description string, url string, urlToImage string, publishedAt time.Time, content string) {

	var insertErr error
	sqlInsert := ""

	log.Println("Initate InsertArticle.")

	// INSERT INTO source (name) VALUES ('Ansa');
	sqlInsert = `INSERT INTO articles (source_id, author, title, description, url, url_to_image,
		published_at, content) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	// Here we use db.Exec as we don't want any record back
	_, insertErr = db.Exec(sqlInsert, sourceID, author, title, description, url, urlToImage, publishedAt, content)

	if insertErr != nil {
		log.Fatal("Error on SQL INSERT => ", insertErr)
	}

	log.Printf("Article stored in the DB.")
}

func main() {

	log.Println("Initiate News collection")

	/* ** get news ** */
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 20)
	results, err := newsapi.FetchNews("", "1", "Italy")

	if err != nil {
		log.Fatal("Error retrieving news => ", err)
	}

	log.Println("News collection completed.")
	log.Printf("Total results retrieved from 'Italy': %v", results.TotalResults)

	log.Println("Iterating on Articles.")

	// Printing News Articles (to be commented)
	/*
		for _, newsArticle := range results.Articles {
			fmt.Printf("Source ID: %v\n", newsArticle.Source.ID)
			fmt.Printf("Source Name: %v\n", newsArticle.Source.Name)
			fmt.Printf("Author: %s\n", newsArticle.Author)
			fmt.Printf("Title: %s\n", newsArticle.Title)
			fmt.Printf("Description: %s\n", newsArticle.Description)
			fmt.Printf("URL: %s\n", newsArticle.URL)
			fmt.Printf("URLToImage: %s\n", newsArticle.URLToImage)
			fmt.Printf("PublishedAt: %v\n", newsArticle.PublishedAt)
			fmt.Printf("Content: %s\n", newsArticle.Content)
		}
	*/
	/* ** write news to db ** */
	myDB := NewDbConnector(db_host, db_port, db_name, db_user, db_password)

	// loop  for each Article
	for i, newsArticle := range results.Articles {
		log.Printf(" >> Article #%d << | Title: '%s'", i+1, newsArticle.Title)
		sourceID := InsertSource(myDB, newsArticle.Source.Name)
		//fmt.Printf("Source ID returned from INSERT: %d\n", sourceID)

		InsertArticle(myDB, sourceID, newsArticle.Author, newsArticle.Title, newsArticle.Description, newsArticle.URL, newsArticle.URLToImage, newsArticle.PublishedAt, newsArticle.Content)

		defer myDB.Close()
	}

	//sourceID := InsertSource(myDB, "Repubblica.it")

	// InsertArticle (myDB, sourceID, newsArticle.Author, newsArticle.Title, newsArticle.Description, newsArticle.URL, newsArticle.URLToImage, newsArticle.PublshedAt, newsArticle.Content)
	// InsertArticle (myDB, sourceID, newsArticle.Author, newsArticle.Title, newsArticle.Description, newsArticle.URL, newsArticle.URLToImage, newsArticle.PublshedAt, newsArticle.Content)

	// --test timestamp --
	/*
		layout := "2021-09-18 16:45:18 +0000 UTC"
		str := "2021-09-18 16:45:18 +0000 UTC"
		t, err := time.Parse(layout, str)
	*/
	/*
		t, err := dateparse.ParseAny("2021-09-18 16:45:18 +0000 UTC")
		if err != nil {
			log.Fatal("Error parsing the Timestamp => ", err)
		}

		fmt.Println("Test Timestamp: ", t)
		InsertArticle(myDB, sourceID, "dummy author", "dummy title", "dummy description", "https://sadsa.xsa", "https://sas.org/image.jpg", t, "dummy content")

		defer myDB.Close()
	*/
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
