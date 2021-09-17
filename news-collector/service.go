package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var environment = getEnv()
var db_host string = "localhost"
var db_port int = 5432
var db_name string = "news"
var db_user string = "news_db_user"
var db_password = environment["db_password"]

func NewDbConnector(db_host string, db_port int, db_name string, db_user string, db_password string) {
	connection_string := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", db_host, db_port, db_user, db_password, db_name)
	//fmt.Println(connection_string)

	// sql.Open() simply validates the arguments provided, doesn't connect yet!
	db_conn, err := sql.Open("postgres", connection_string)
	if err != nil {
		panic(err)
	}

	defer db_conn.Close()

	// the method Ping() is actually attempting a connection to the database
	err = db_conn.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("DB Connection Successful!")
}

func main() {

	NewDbConnector(db_host, db_port, db_name, db_user, db_password)

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
