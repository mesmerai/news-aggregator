package data

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type DBClient struct {
	Database *sql.DB
}

func NewDBClient(db_host string, db_port int, db_name, db_user, db_password string, maxRetries int) (db *DBClient) {

	log.Println("Initiate Connection to DB.")

	// currently 'sslmode=verify-full' gives error
	connection_string := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", db_host, db_port, db_user, db_password, db_name)

	// sql.Open() simply validates the arguments provided, doesn't connect yet!
	db_conn, err := sql.Open("postgres", connection_string)
	if err != nil {
		log.Fatal("Error validating DB connection parameters => ", err)
	}

	// the method Ping() is actually attempting a connection to the database
	for i := 0; i < maxRetries; i++ {
		err = db_conn.Ping()
		if err == nil {
			break
		}
		log.Println("DB connection attempt #", i+1)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatal("Error connecting to DB => ", err)
	}

	log.Println("Connection to DB successful.")
	return &DBClient{db_conn}

}

func (db *DBClient) GetFavourites() *sql.Rows {

	log.Printf("Initiate GetFavourites")

	var selectRows *sql.Rows
	var selectErr error
	sqlSelect := ""

	sqlSelect = "SELECT * FROM domains WHERE favourite = TRUE"

	selectRows, selectErr = db.Database.Query(sqlSelect)
	if selectErr != nil {
		log.Fatal("Error on SQL SELECT => ", selectErr)
	}

	return selectRows

}

func (db *DBClient) GetDomainID(name string) int {

	log.Printf("Initiate GetDomainID")

	var id int
	var selectRow *sql.Row
	var selectErr error
	sqlSelect := ""

	sqlSelect = "SELECT id FROM domains WHERE name = $1"

	selectRow = db.Database.QueryRow(sqlSelect, name)
	selectErr = selectRow.Scan(&id)
	if selectErr != nil {
		log.Fatal("Error on SQL SELECT => ", selectErr)
	}

	log.Printf("Domain ID for '%s' is: ", name)

	return id

}

// ** THIS METHOD SHOULD NOT BE NEEDED ANYMORE. REPLACED BY GetFavourites() **
// ** HOWEVER i WILL LEAVE IT HERE AS IT'S A COOL CODE BLOCK **
/*
func (db *DBClient) GetDomains(filter []string) *sql.Rows {

	log.Printf("Initiate GetDomains")

	var selectRows *sql.Rows
	var selectErr error
	//sqlSelect := ""

	buf := bytes.NewBufferString("SELECT * FROM domains WHERE name in ('")
	for i, dName := range filter {
		if i > 0 {
			buf.WriteString("','")
		}
		buf.WriteString(dName)
	}
	buf.WriteString("')")

	fmt.Println("SQL: ", buf.String())

	selectRows, selectErr = db.Database.Query(buf.String())
	if selectErr != nil {
		log.Fatal("Error on SQL SELECT => ", selectErr)
	}

	return selectRows

}
*/

func (db *DBClient) GetDomainsByName(name string) *sql.Rows {

	log.Printf("Initiate GetDomainsByName")

	var selectRows *sql.Rows
	var selectErr error
	sqlSelect := ""

	sqlSelect = "SELECT * FROM domains WHERE name = $1"

	selectRows, selectErr = db.Database.Query(sqlSelect, name)
	if selectErr != nil {
		log.Fatal("Error on SQL SELECT => ", selectErr)
	}

	return selectRows

}

func (db *DBClient) GetSourcesByName(name string) *sql.Rows {

	log.Printf("Initiate GetSourcsByName")

	var selectRows *sql.Rows
	var sqlSelect = ""
	var selectErr error

	sqlSelect = "SELECT * FROM sources WHERE name = $1"

	selectRows, selectErr = db.Database.Query(sqlSelect, name)
	if selectErr != nil {
		log.Fatal("Error on SQL SELECT => ", selectErr)
	}

	return selectRows

}

func (db *DBClient) InsertSource(sourceName string) (sourceID int) {

	log.Printf("Initiate InsertSource for %s", sourceName)

	id := 0
	var insertRow *sql.Row
	var insertErr error

	// INSERT INTO source (name) VALUES ('Ansa');
	var sqlInsert = `INSERT INTO sources (name) VALUES ($1) RETURNING id`

	// QueryRow returns a *Row
	insertRow = db.Database.QueryRow(sqlInsert, sourceName)
	insertErr = insertRow.Scan(&id)
	if insertErr != nil {
		log.Fatal("Error on SQL INSERT => ", insertErr)
	}

	log.Printf("Source '%s' stored in the DB.", sourceName)

	// return the id of the source just INSERTed
	return id
}

func (db *DBClient) InsertDomain(domainName string) (domainID int) {

	log.Printf("Initiate InsertDomain for %s", domainName)

	id := 0
	var insertRow *sql.Row
	var insertErr error
	sqlInsert := ""

	// INSERT INTO source (name) VALUES ('Ansa');
	sqlInsert = `INSERT INTO domains (name) VALUES ($1) RETURNING id`

	// QueryRow returns a *Row
	insertRow = db.Database.QueryRow(sqlInsert, domainName)
	insertErr = insertRow.Scan(&id)
	if insertErr != nil {
		log.Fatal("Error on SQL INSERT => ", insertErr)
	}

	log.Printf("Domain '%s' stored in the DB.", domainName)

	// return the id of the source just INSERTed
	return id

}

func (db *DBClient) InsertArticle(sourceID, domainID int, author string, title string, description string, url string, urlToImage string, publishedAt time.Time, content, country, language, category string) {

	var insertErr error
	sqlInsert := ""

	log.Println("Initate InsertArticle.")

	// INSERT INTO source (name) VALUES ('Ansa');
	sqlInsert = `INSERT INTO articles (source_id, domain_id, author, title, description, url, url_to_image,
		published_at, content, country, language, category) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	// Here we use db.Exec as we don't want any record back
	_, insertErr = db.Database.Exec(sqlInsert, sourceID, domainID, author, title, description, url, urlToImage, publishedAt, content, country, language, category)

	if insertErr != nil {
		log.Fatal("Error on SQL INSERT => ", insertErr)
	}

	log.Printf("Article stored in the DB.")
}
