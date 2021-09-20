package main

import (
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/mesmerai/news-aggregator/ncollector/data"
	"github.com/mesmerai/news-aggregator/ncollector/news"
)

var environment = getEnv()
var db_host string = "localhost"
var db_port int = 5432
var db_name string = "news"
var db_user string = "news_db_user"

// from Env
var db_password = environment["db_password"]
var news_api_key string = environment["news_api_key"]

func main() {

	/* ** get news ** */
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 100)

	FetchAndStore(newsapi, "ByCountry", "Italy")
	FetchAndStore(newsapi, "ByCountry", "Australia")
	//FetchAndStore(newsapi, "Global", "")
	// Global search need to be restricted by some parameter (e.g from/to, source ..)

}

func FetchAndStore(n *news.Client, searchType, searchParameter string) {

	var results *news.Results
	var err error

	log.Printf("Initiate News collection: %s", searchType)

	switch {
	case searchType == "ByCountry":
		if searchParameter == "" {
			log.Fatal("searchParameter must specified in SearchByCountry")
		} else {
			results, err = n.FetchNews("ByCountry", "", "1", "Italy")
		}
	case searchType == "Global":
		results, err = n.FetchNews("Global", "", "1", "")
	default:
		log.Fatal("Search type Must be specified. Allowed values: 'Global', 'ByCountry'")
	}

	if err != nil {
		log.Fatal("Error retrieving news => ", err)
	}

	log.Println("News collection completed.")
	log.Printf("Total results retrieved from 'Italy': %v", results.TotalResults)

	log.Println("Iterating on Articles.")

	/* ** write news to db ** */
	myDB := data.NewDbConnector(db_host, db_port, db_name, db_user, db_password)

	// loop  for each Article
	for i, newsArticle := range results.Articles {
		log.Printf(" >> Article #%d << | Title: '%s'", i+1, newsArticle.Title)
		sourceID := data.InsertSource(myDB, newsArticle.Source.Name)
		//fmt.Printf("Source ID returned from INSERT: %d\n", sourceID)

		data.InsertArticle(myDB, sourceID, newsArticle.Author, newsArticle.Title, newsArticle.Description, newsArticle.URL, newsArticle.URLToImage, newsArticle.PublishedAt, newsArticle.Content)

		defer myDB.Close()
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
