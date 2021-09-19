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

	/*
		var thisID int
		var thisName string
	*/

	// check if Source exists
	sqlSelect = "SELECT * FROM source WHERE name = $1"
	log.Printf("Retrieving sources for '%s'", sourceName)
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

		fmt.Printf("- rows.id: %d\n", thisSource.id)
		fmt.Printf("- rows.name: %s\n", thisSource.name)

	}

	defer selectRows.Close()

	/*
		fmt.Println(" -- Printing sources slice --")
		fmt.Printf("source slice: %v\n", sources)
		fmt.Printf("len(source): %d\n", len(sources))
	*/

	// the 'source' slide is empty if no rows are returned by the SELECT
	if len(sources) == 0 {
		log.Println("No sources found. Proceed with INSERT.")

		// INSERT INTO source (name) VALUES ('Ansa');
		sqlInsert = `INSERT INTO source (name) VALUES ($1) RETURNING id`

		// QueryRow returns a *Row
		insertRow = db.QueryRow(sqlInsert, sourceName)
		insertErr = insertRow.Scan(&id)
		if insertErr != nil {
			log.Fatal("Error on SQL INSERT => ", insertErr)
		}

		// return the id of the source just INSERTed
		return id

	} else {

		// return the id of existing source, so that we can link as foreign key in Article
		return thisSource.id
	}

}

func main() {

	// connect to DB
	myDB := NewDbConnector(db_host, db_port, db_name, db_user, db_password)

	sourceID := InsertSource(myDB, "Repubblica.it")
	fmt.Printf("Source ID returned from INSERT: %d\n", sourceID)

	defer myDB.Close()

	// get news
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 20)
	results, err := newsapi.FetchNews("", "1", "Italy")

	if err != nil {
		log.Fatal("Error retrieving news => ", err)
	}

	fmt.Println(" ")
	fmt.Println("-- News Total Results --")
	fmt.Println(results.TotalResults)
	fmt.Println("-- Iterate on each Article --")
	for _, this_article := range results.Articles {
		fmt.Printf("Source ID: %v\n", this_article.Source.ID)
		fmt.Printf("Source Name: %v\n", this_article.Source.Name)
		fmt.Printf("Title: %s\n", this_article.Title)
	}

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
