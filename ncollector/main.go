package main

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
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

	//newsapi.FetchSources()

	FetchAndStoreArticles(newsapi, "ByCountry", "Italy")
	FetchAndStoreArticles(newsapi, "ByCountry", "Australia")

	// here should call a Global Search for one specific domain to test
	// --> this need to loop for each domain stored in the DB
	FetchAndStoreArticles(newsapi, "Global", "news.com.au")

	/* ** Set time window for Global Search ** */
	// from 1 hour to now
	/*
		to := time.Now()
		fmt.Printf("To: %s\n", to)

		from := to.Add(-1 * time.Hour)
		fmt.Printf("From: %s\n", from)

		toRFC := to.Format(time.RFC3339)
		fmt.Printf("To (RFC): %s\n", toRFC)

		fromRFC := from.Format(time.RFC3339)
		fmt.Printf("From (RFC): %s\n", fromRFC)
	*/
	// Global search Must be restricted with one of those params (q, qInTitle, sources, domains) otherwise doesn't work
	//FetchAndStoreArticles(newsapi, "Global", "", fromRFC, toRFC)

}

func FetchAndStoreSources() {
	// to do (maybe not)
}

func FetchAndStoreArticles(n *news.Client, searchType, searchParameter string) {

	var results *news.Results
	var err error

	log.Println("==========================================================")
	log.Printf("Initiate News collection: %s", searchType)
	log.Println("==========================================================")

	switch {
	case searchType == "ByCountry":
		if searchParameter == "" {
			log.Fatal("searchParameter must be set with the country name in SearchByCountry")
		} else {
			results, err = n.FetchNews("ByCountry", "", "1", searchParameter)
		}
	case searchType == "Global":
		if searchParameter == "" {
			log.Fatal("searchParameter must be set with the domain name in GlobalSearch")
		} else {
			results, err = n.FetchNews("Global", "", "1", searchParameter)
		}
	default:
		log.Fatal("Search type Must be specified. Allowed values: 'Global', 'ByCountry'")
	}

	if err != nil {
		log.Fatal("Error retrieving news => ", err)
	}

	log.Println("News collection completed.")
	log.Printf("Total results retrieved for '%s': %v", searchParameter, results.TotalResults)

	/* ** write news to db ** */
	myDB := data.NewDbConnector(db_host, db_port, db_name, db_user, db_password)

	log.Println("--------------------------------------------------------")
	log.Println("Iterating on Articles.")
	log.Println("--------------------------------------------------------")

	// loop  for each Article and call:
	// insertSource, insertDomain, insertArticle
	for i, newsArticle := range results.Articles {
		log.Printf(" >> Article #%d << | Title: '%s'", i+1, newsArticle.Title)

		/* ** insertSource ** */
		sourceID := data.InsertSource(myDB, newsArticle.Source.Name)
		//fmt.Printf("Source ID returned from INSERT: %d\n", sourceID)

		/* ** insertDomain ** */
		// extract domain name from article URL. From https//www.techcrunch.com/zyx TO techcrunch.com
		log.Println("URL: ", newsArticle.URL)

		// cut the http(s) and www part from URL, if present
		regexp := regexp.MustCompile(`http(s?)://(www.)?`)
		cutURL := regexp.ReplaceAllString(newsArticle.URL, "")

		components := strings.Split(cutURL, "/")
		domain := components[0]
		log.Println("Domain extracted from URL: ", domain)

		// then call
		domainID := data.InsertDomain(myDB, domain)

		data.InsertArticle(myDB, sourceID, domainID, newsArticle.Author, newsArticle.Title, newsArticle.Description, newsArticle.URL, newsArticle.URLToImage, newsArticle.PublishedAt, newsArticle.Content, "", "", "")

		defer myDB.Close()
		log.Println("Closing DB resources.")
		log.Println("--------------------------------------------------------")

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
