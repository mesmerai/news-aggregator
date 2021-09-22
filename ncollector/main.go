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
	id   int
	name string
}

var thisDomain = Domain{}

func main() {

	/* ** News Client ** */
	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, news_api_key, 100)

	/* ** DB Conn ** */
	myDB := data.NewDBClient(db_host, db_port, db_name, db_user, db_password)

	// myDB = *DBClient(db_conn)
	defer myDB.Database.Close()

	log.Println("Closing DB resources.")

	log.Println("==========================================================")
	log.Println("Initiate News collection")
	log.Println("==========================================================")

	CountryFetchAndStore(myDB, newsapi, "Italy", "Italian")
	CountryFetchAndStore(myDB, newsapi, "Australia", "English")
	GlobalFetchAndStore(myDB, newsapi)

}

func GlobalFetchAndStore(myDB *data.DBClient, newsapi *news.Client) {

	domainRows := myDB.GetDomains()

	for domainRows.Next() {

		// Scan copies the columns in the current row into the values pointed at by dest.
		// The number of values in dest must be the same as the number of columns in Rows.
		err := domainRows.Scan(&thisDomain.id, &thisDomain.name)
		if err != nil {
			log.Fatal("Error on reading SQL SELECT results => ", err)
		}

		log.Println("**********************************************************")
		log.Println("Global Search ByDomain: ", thisDomain.name)
		log.Println("**********************************************************")

		results, err := newsapi.FetchNews("Global", "", "1", thisDomain.name)
		if err != nil {
			log.Fatal("Error retrieving news => ", err)
		}

		log.Println("News collection completed.")
		log.Printf("Total results retrieved for '%s': %v", thisDomain.name, results.TotalResults)

		log.Println("--------------------------------------------------------")
		log.Println("Iterating on Articles.")
		log.Println("--------------------------------------------------------")

		for i, newsArticle := range results.Articles {
			log.Printf(" >> Article #%d << | Title: '%s'", i+1, newsArticle.Title)

			sources := make([]string, 0)

			// exists for sure
			domainID := myDB.GetDomainID(thisDomain.name)

			/* ** Check Sources ** */
			sourceRows := myDB.GetSourcesByName(newsArticle.Source.Name)

			for sourceRows.Next() {

				// Scan copies the columns in the current row into the values pointed at by dest.
				// The number of values in dest must be the same as the number of columns in Rows.
				err := sourceRows.Scan(&thisSource.id, &thisSource.name)
				if err != nil {
					log.Fatal("Error on reading SQL SELECT results => ", err)
				}
				// fill the slice to check if there are rows later
				sources = append(sources, thisSource.name)
				log.Printf("Records found for source: '%s'", newsArticle.Source.Name)
				log.Printf(" - rows.id: %d\n", thisSource.id)
				log.Printf(" - rows.name: %s\n", thisSource.name)

			}
			log.Println("Closing rows resources.")
			defer sourceRows.Close()

			// the 'source' slide is empty if no rows are returned by the SELECT
			if len(sources) == 0 {
				log.Printf("No sources found for '%s'. Proceed with INSERT.", newsArticle.Source.Name)

				/* ** insertSource ** */
				sourceID = myDB.InsertSource(newsArticle.Source.Name)
			} else {
				log.Println("The source already exists. No INSERT required.")
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
	log.Println("Search ByCountry: ", country)
	log.Println("**********************************************************")

	// -- potentially can call a function
	// -- e.g. for Italy (require iternation of Articles and domain regexp defined)
	// CheckAdStore (DB, country, source, domain) ?!
	// CheckAndStore (myDB, "Italy", newsArticle.Source.Name, domain)

	results, err := newsapi.FetchNews("ByCountry", "", "1", country)
	if err != nil {
		log.Fatal("Error retrieving news => ", err)
	}

	log.Println("News collection completed.")
	log.Printf("Total results retrieved for '%s': %v", country, results.TotalResults)

	log.Println("--------------------------------------------------------")
	log.Println("Iterating on Articles.")
	log.Println("--------------------------------------------------------")

	for i, newsArticle := range results.Articles {
		log.Printf(" >> Article #%d << | Title: '%s'", i+1, newsArticle.Title)

		// there's nothing provided in the Sql package to check if Rows has no records inside
		// so I define an empty slice and fill it in the iteration below
		sources := make([]string, 0)
		domains := make([]string, 0)

		/* ** Extract Domain ** */
		// extract domain name from article URL. From https//www.techcrunch.com/zyx TO techcrunch.com
		log.Println("URL: ", newsArticle.URL)
		// cut the http(s) and www part from URL, if present
		regexp := regexp.MustCompile(`http(s?)://(www.)?`)
		cutURL := regexp.ReplaceAllString(newsArticle.URL, "")
		components := strings.Split(cutURL, "/")
		domain := components[0]
		log.Println("Domain extracted from URL: ", domain)

		/* ** Check Domains ** */
		domainRows := myDB.GetDomainsByName(domain)

		for domainRows.Next() {

			// Scan copies the columns in the current row into the values pointed at by dest.
			// The number of values in dest must be the same as the number of columns in Rows.
			err := domainRows.Scan(&thisDomain.id, &thisDomain.name)
			if err != nil {
				log.Fatal("Error on reading SQL SELECT results => ", err)
			}
			// fill the slice to check if there are rows later
			domains = append(domains, thisDomain.name)
			log.Printf("Records found for domain: '%s'", domain)
			log.Printf(" - rows.id: %d\n", thisDomain.id)
			log.Printf(" - rows.name: %s\n", thisDomain.name)

		}
		log.Println("Closing rows resources.")
		defer domainRows.Close()

		// the 'source' slide is empty if no rows are returned by the SELECT
		if len(domains) == 0 {
			log.Printf("No domains found for '%s'. Proceed with INSERT.", domain)

			/* ** insertDomain ** */
			domainID = myDB.InsertDomain(domain)
		} else {
			log.Println("The domain already exists. No INSERT required.")
			domainID = thisDomain.id
		}

		/* ** Check Sources ** */
		sourceRows := myDB.GetSourcesByName(newsArticle.Source.Name)

		for sourceRows.Next() {

			// Scan copies the columns in the current row into the values pointed at by dest.
			// The number of values in dest must be the same as the number of columns in Rows.
			err := sourceRows.Scan(&thisSource.id, &thisSource.name)
			if err != nil {
				log.Fatal("Error on reading SQL SELECT results => ", err)
			}
			// fill the slice to check if there are rows later
			sources = append(sources, thisSource.name)
			log.Printf("Records found for source: '%s'", newsArticle.Source.Name)
			log.Printf(" - rows.id: %d\n", thisSource.id)
			log.Printf(" - rows.name: %s\n", thisSource.name)

		}
		log.Println("Closing rows resources.")
		defer sourceRows.Close()

		// the 'source' slide is empty if no rows are returned by the SELECT
		if len(sources) == 0 {
			log.Printf("No sources found for '%s'. Proceed with INSERT.", newsArticle.Source.Name)

			/* ** insertSource ** */
			sourceID = myDB.InsertSource(newsArticle.Source.Name)
		} else {
			log.Println("The source already exists. No INSERT required.")
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
	db_password := os.Getenv("DB_PASSWORD")
	if db_password == "" {
		log.Fatal("Password for the DB is not set in ENV.")
	}

	envMap["news_api_key"] = news_api_key
	envMap["db_password"] = db_password

	return envMap
}
