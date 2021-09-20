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

	log.Println("Initiate News collection")

	/* ** get news ** */
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 100)
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
	myDB := data.NewDbConnector(db_host, db_port, db_name, db_user, db_password)

	// loop  for each Article
	for i, newsArticle := range results.Articles {
		log.Printf(" >> Article #%d << | Title: '%s'", i+1, newsArticle.Title)
		sourceID := data.InsertSource(myDB, newsArticle.Source.Name)
		//fmt.Printf("Source ID returned from INSERT: %d\n", sourceID)

		data.InsertArticle(myDB, sourceID, newsArticle.Author, newsArticle.Title, newsArticle.Description, newsArticle.URL, newsArticle.URLToImage, newsArticle.PublishedAt, newsArticle.Content)

		defer myDB.Close()
	}

	// manual inserts
	//sourceID := InsertSource(myDB, "Repubblica.it")
	//InsertArticle(myDB, sourceID, "dummy author", "dummy title", "dummy description", "https://sadsa.xsa", "https://sas.org/image.jpg", t, "dummy content")

	// --test timestamp native --
	/*
		layout := "2021-09-18 16:45:18 +0000 UTC"
		str := "2021-09-18 16:45:18 +0000 UTC"
		t, err := time.Parse(layout, str)
	*/
	/* -- test timestamp github.com/araddon/dateparse --
	t, err := dateparse.ParseAny("2021-09-18 16:45:18 +0000 UTC")
	if err != nil {
		log.Fatal("Error parsing the Timestamp => ", err)
	}

	fmt.Println("Test Timestamp: ", t)

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
