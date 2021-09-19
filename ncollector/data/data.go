package data

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

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
