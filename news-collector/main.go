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

func main() {

	// connect to DB
	myDB := NewDbConnector(db_host, db_port, db_name, db_user, db_password)

	// INSERT INTO source (name) VALUES ('Ansa');
	sqlInsert := `INSERT INTO source (name) VALUES ($1)`
	_, err := myDB.Exec(sqlInsert, "Ansa.it")
	if err != nil {
		log.Fatal("Error on SQL INSERT => ", err)
	}

	defer myDB.Close()

	// get news
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 20)
	results, err := newsapi.FetchNews("", "1", "Italy")

	if err != nil {
		log.Fatal("Error retrieving news => ", err)
	}

	fmt.Println("-- Total Results --")
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
