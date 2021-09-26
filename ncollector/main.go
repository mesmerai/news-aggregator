package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/mileusna/crontab"

	"github.com/mesmerai/news-aggregator/ncollector/data"
	"github.com/mesmerai/news-aggregator/ncollector/news"
)

var environment = getEnv()

var db_port int = 5432
var db_name string = "news"
var db_user string = "news_db_user"
var dbconn_max_retries = 10

// from Env
var news_api_key string = environment["news_api_key"]
var db_host = environment["db_host"]
var db_password = environment["db_password"]

/* ** vars ** */
var sourceID int
var domainID int

// Source struct to deal with INSERT later. Declare and initialize.
type Source struct {
	id   int
	name string
}

var thisSource = Source{}

// Domain struct to deal with INSERT later. Declare and initialize.
type Domain struct {
	id        int
	name      string
	favourite bool
}

var thisDomain = Domain{}

type FavouriteFeed struct {
	id        int
	name      string
	favourite bool
}

var thisFeed = FavouriteFeed{}

func main() {

	// ** Entire Block Schedule to run every 3 hours **
	//
	// MAX 25 API Calls in 6 hours - 23 Max Feeds
	// MAX 12 API Calls in 3 hours - 10 Max Feeds
	log.Println("Initiating Cron Jobs")
	ctab := crontab.New() // create cron table

	//ctab.MustAddJob("* * * * *", FetchItaly)
	// Run every 3 hours - Does not work, run every minute!
	//ctab.MustAddJob("* */3 * * *", FetchItaly)
	//ctab.MustAddJob("* */3 * * *", FetchAustralia)
	//ctab.MustAddJob("* */3 * * *", FetchGlobal)

	// run at 9:10, 12:10, 15:10 etc..
	ctab.MustAddJob("10 9,12,15 * * *", FetchItaly)
	ctab.MustAddJob("15 9,12,15 * * *", FetchAustralia)
	ctab.MustAddJob("20 9,12,15 * * *", FetchGlobal)

	// troubleshooting: run every 30 minutes
	//ctab.MustAddJob("*/30 * * * *", FetchItaly)
	//ctab.MustAddJob("*/30 * * * *", FetchAustralia)
	//ctab.MustAddJob("*/30 * * * *", FetchGlobal)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	//srv.db.close()
	ctab.Clear()
	log.Println("Clear Cron Resources")
}

func FetchGlobal() {

	log.Println("==========================================================")
	log.Println("Global | News Collection Start")
	log.Println("==========================================================")

	/* ** News Client ** */
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 100)

	/* ** DB Conn ** */
	myDB := data.NewDBClient(db_host, db_port, db_name, db_user, db_password, dbconn_max_retries)

	// myDB = *DBClient(db_conn)
	defer myDB.Database.Close()

	log.Println("Global | Closing DB resources.")

	// restricted list of domains REQUIRED to not reach the API call daily LIMIT of 50 API calls in 12 hours
	//dList := []string{"corriere.it", "ansa.it", "rainews.it"}

	// ** COMMENTING BEFORE DELETE **
	/*
		rows := myDB.GetFavourites()
		dList := make([]string, 0)

		for rows.Next() {
			err := rows.Scan(&thisDomain.id, &thisDomain.name, &thisDomain.favourite)
			if err != nil {
				log.Fatal("Error on reading SQL SELECT results => ", err)
			}

			dList = append(dList, thisDomain.name)
		}

		log.Println("Global | Favourite Feeds: ", dList)
		GlobalFetchAndStore(myDB, newsapi, dList)
	*/

	GlobalFetchAndStore(myDB, newsapi)

	log.Println("Global | News Collection End")

}

func FetchItaly() {

	log.Println("==========================================================")
	log.Println("ByCountry | News Collection Start")
	log.Println("==========================================================")

	/* ** News Client ** */
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 100)

	/* ** DB Conn ** */
	myDB := data.NewDBClient(db_host, db_port, db_name, db_user, db_password, dbconn_max_retries)

	// myDB = *DBClient(db_conn)
	defer myDB.Database.Close()

	log.Println("ByCountry | Closing DB resources.")

	CountryFetchAndStore(myDB, newsapi, "Italy", "Italian")

	log.Println("ByCountry | News Collection End")

}

func FetchAustralia() {

	log.Println("==========================================================")
	log.Println("ByCountry | News Collection Start")
	log.Println("==========================================================")

	/* ** News Client ** */
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 100)

	/* ** DB Conn ** */
	myDB := data.NewDBClient(db_host, db_port, db_name, db_user, db_password, dbconn_max_retries)

	// myDB = *DBClient(db_conn)
	defer myDB.Database.Close()

	log.Println("ByCountry | Closing DB resources.")

	CountryFetchAndStore(myDB, newsapi, "Australia", "English")

	log.Println("ByCountry | News Collection End")

}

// API call for each domain - LIMIT per Dev plan reached at 50 calls in 12 hours
func GlobalFetchAndStore(myDB *data.DBClient, newsapi *news.Client) {

	// ** COMMENTING BEFORE DELETE **
	/*
		domainRows := myDB.GetDomains(domainsList)
	*/

	feedRows := myDB.GetFavourites()

	for feedRows.Next() {

		// Scan copies the columns in the current row into the values pointed at by dest.
		// The number of values in dest must be the same as the number of columns in Rows.
		err := feedRows.Scan(&thisFeed.id, &thisFeed.name, &thisFeed.favourite)
		if err != nil {
			log.Fatal("Global | Error on reading SQL SELECT results => ", err)
		}

		log.Println("**********************************************************")
		log.Println("Global | Search ByDomain: ", thisFeed.name)
		log.Println("**********************************************************")

		results, err := newsapi.FetchNews("Global", "", "1", thisFeed.name)
		if err != nil {
			log.Fatal("Global | Error retrieving news => ", err)
		}

		log.Printf("Global | Total results retrieved for '%s': %v", thisFeed.name, results.TotalResults)

		log.Println("--------------------------------------------------------")
		log.Println("Global | Iterating on Articles.")
		log.Println("--------------------------------------------------------")

		for i, newsArticle := range results.Articles {
			log.Printf("Global | Article #%d | Title: '%s'", i+1, newsArticle.Title)

			sources := make([]string, 0)

			// exists for sure
			domainID := myDB.GetDomainID(thisFeed.name)

			/* ** Check Sources ** */
			sourceRows := myDB.GetSourcesByName(newsArticle.Source.Name)

			for sourceRows.Next() {

				// Scan copies the columns in the current row into the values pointed at by dest.
				// The number of values in dest must be the same as the number of columns in Rows.
				err := sourceRows.Scan(&thisSource.id, &thisSource.name)
				if err != nil {
					log.Fatal("Global | Error on reading SQL SELECT results => ", err)
				}
				// fill the slice to check if there are rows later
				sources = append(sources, thisSource.name)
				log.Printf("Global | Records found for source: '%s'", newsArticle.Source.Name)
				log.Printf("Global |  - rows.id: %d\n", thisSource.id)
				log.Printf("Global |  - rows.name: %s\n", thisSource.name)

			}
			log.Println("Global | Closing rows resources.")
			defer sourceRows.Close()

			// the 'source' slide is empty if no rows are returned by the SELECT
			if len(sources) == 0 {
				log.Printf("Global | No sources found for '%s'. Proceed with INSERT.", newsArticle.Source.Name)

				/* ** insertSource ** */
				sourceID = myDB.InsertSource(newsArticle.Source.Name)
			} else {
				log.Println("Global | The source already exists. No INSERT required.")
				sourceID = thisSource.id
			}

			/* ** InsertArticle ** */
			myDB.InsertArticle(sourceID, domainID, newsArticle.Author, newsArticle.Title, newsArticle.Description, newsArticle.URL, newsArticle.URLToImage, newsArticle.PublishedAt, newsArticle.Content, "", "", "")

			log.Println("--------------------------------------------------------")

		}

	}

}

func CountryFetchAndStore(myDB *data.DBClient, newsapi *news.Client, country, language string) {

	/* ********** Start with Italy ***************************************** */
	log.Println("**********************************************************")
	log.Println("ByCountry | Search :", country)
	log.Println("**********************************************************")

	// -- potentially can call a function
	// -- e.g. for Italy (require iternation of Articles and domain regexp defined)
	// CheckAdStore (DB, country, source, domain) ?!
	// CheckAndStore (myDB, "Italy", newsArticle.Source.Name, domain)

	results, err := newsapi.FetchNews("ByCountry", "", "1", country)
	if err != nil {
		log.Fatal("ByCountry | Error retrieving news => ", err)
	}

	log.Printf("ByCountry | Total results retrieved for '%s': %v", country, results.TotalResults)

	log.Println("--------------------------------------------------------")
	log.Println("ByCountry | Iterating on Articles.")
	log.Println("--------------------------------------------------------")

	for i, newsArticle := range results.Articles {
		log.Printf("ByCountry |  Article #%d | Title: '%s'", i+1, newsArticle.Title)

		// there's nothing provided in the Sql package to check if Rows has no records inside
		// so I define an empty slice and fill it in the iteration below
		sources := make([]string, 0)
		domains := make([]string, 0)

		/* ** Extract Domain ** */
		// extract domain name from article URL. From https//www.techcrunch.com/zyx TO techcrunch.com
		log.Println("ByCountry | URL: ", newsArticle.URL)
		// cut the http(s) and www part from URL, if present
		regexp := regexp.MustCompile(`http(s?)://(www.)?`)
		cutURL := regexp.ReplaceAllString(newsArticle.URL, "")
		components := strings.Split(cutURL, "/")
		domain := components[0]
		log.Println("ByCountry | Domain extracted from URL: ", domain)

		/* ** Check Domains ** */
		domainRows := myDB.GetDomainsByName(domain)

		for domainRows.Next() {

			// Scan copies the columns in the current row into the values pointed at by dest.
			// The number of values in dest must be the same as the number of columns in Rows.
			err := domainRows.Scan(&thisDomain.id, &thisDomain.name, &thisDomain.favourite)
			if err != nil {
				log.Fatal("ByCountry | Error on reading SQL SELECT results => ", err)
			}
			// fill the slice to check if there are rows later
			domains = append(domains, thisDomain.name)
			log.Printf("ByCountry | Records found for domain: '%s'", domain)
			log.Printf("ByCountry |  - rows.id: %d\n", thisDomain.id)
			log.Printf("ByCountry |  - rows.name: %s\n", thisDomain.name)

		}
		log.Println("ByCountry | Closing rows resources.")
		defer domainRows.Close()

		// the 'source' slide is empty if no rows are returned by the SELECT
		if len(domains) == 0 {
			log.Printf("ByCountry | No domains found for '%s'. Proceed with INSERT.", domain)

			/* ** insertDomain ** */
			domainID = myDB.InsertDomain(domain)
		} else {
			log.Println("ByCountry | The domain already exists. No INSERT required.")
			domainID = thisDomain.id
		}

		/* ** Check Sources ** */
		sourceRows := myDB.GetSourcesByName(newsArticle.Source.Name)

		for sourceRows.Next() {

			// Scan copies the columns in the current row into the values pointed at by dest.
			// The number of values in dest must be the same as the number of columns in Rows.
			err := sourceRows.Scan(&thisSource.id, &thisSource.name)
			if err != nil {
				log.Fatal("ByCountry | Error on reading SQL SELECT results => ", err)
			}
			// fill the slice to check if there are rows later
			sources = append(sources, thisSource.name)
			log.Printf("ByCountry | Records found for source: '%s'", newsArticle.Source.Name)
			log.Printf("ByCountry |  - rows.id: %d\n", thisSource.id)
			log.Printf("ByCountry |  - rows.name: %s\n", thisSource.name)

		}
		log.Println("ByCountry | Closing rows resources.")
		defer sourceRows.Close()

		// the 'source' slide is empty if no rows are returned by the SELECT
		if len(sources) == 0 {
			log.Printf("ByCountry | No sources found for '%s'. Proceed with INSERT.", newsArticle.Source.Name)

			/* ** insertSource ** */
			sourceID = myDB.InsertSource(newsArticle.Source.Name)
		} else {
			log.Println("ByCountry | The source already exists. No INSERT required.")
			sourceID = thisSource.id
		}

		/* ** InsertArticle ** */
		myDB.InsertArticle(sourceID, domainID, newsArticle.Author, newsArticle.Title, newsArticle.Description, newsArticle.URL, newsArticle.URLToImage, newsArticle.PublishedAt, newsArticle.Content, country, language, "")

		log.Println("--------------------------------------------------------")

	}

}

func getEnv() map[string]string {

	envMap := map[string]string{}

	news_api_key := os.Getenv("NEWS_API_KEY")
	if news_api_key == "" {
		log.Fatal("News Api Key is not set in ENV.")
	}
	db_host := os.Getenv("DB_HOST")
	if db_host == "" {
		log.Fatal("DB_HOST is not set in ENV.")
	}
	db_password := os.Getenv("DB_PASSWORD")
	if db_password == "" {
		log.Fatal("Password for the DB is not set in ENV.")
	}

	envMap["news_api_key"] = news_api_key
	envMap["db_host"] = db_host
	envMap["db_password"] = db_password

	return envMap
}
