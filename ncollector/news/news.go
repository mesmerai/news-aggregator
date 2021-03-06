package news

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

/* Article structs */
type Article struct {
	Source struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

/* Source struct (not used yet) */
type Source struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Category    string `json:"category"`
	Language    string `json:"language"`
	Country     string `json:"country"`
}

type Sources struct {
	Status  string   `json:"status"`
	Sources []Source `json:"sources"`
}

// Countries Map
var countries = map[string]string{
	"Australia": "au",
	"Italy":     "it",
}

/* Client is our struct for the News Client
- http is a pointer to the httpClient itself that makes the web requests
- key is the API key
- PageSize is the number of results to return per page (max 100)
*/
// IMPORTANT: only PageSize is Exported to other packages as it's CAPITAL Letter
type Client struct {
	http     *http.Client
	key      string
	PageSize int
}

// NewClient function creates our Client used for requests
func NewClient(httpClient *http.Client, key string, pageSize int) *Client {
	if pageSize > 100 {
		pageSize = 100
	}

	return &Client{httpClient, key, pageSize}
}

// format the 'PublishedAt' date
func (a *Article) FormatPublishedDate() string {
	// func (t Time) Date() (year int, month Month, day int)
	year, month, day := a.PublishedAt.Date()
	// %v	the value in a default format
	return fmt.Sprintf("%v %d, %d", month, day, year)

}

func (c *Client) FetchSources() {

	var endpoint = ""

	//https://newsapi.org/v2/top-headlines/sources?apiKey=API_KEY
	endpoint = fmt.Sprintf("https://newsapi.org/v2/top-headlines/sources?apiKey=%s", c.key)

	resp, err := c.http.Get(endpoint)

	// Handle error from the response
	if err != nil {
		log.Fatal("Error getting a response => ", err)
	}

	defer resp.Body.Close()

	// Response body is converted to a byte slice, if no errors
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Error reading the response => ", err)
	}

	// this is for Printing Response Body and Ret Code
	fmt.Println(string(body))
	fmt.Printf("Response Status: %s\n", resp.Status)

	// checking ret code, http.StatusOk is a const from http pkg
	if resp.StatusCode != http.StatusOK {
		log.Fatal("Error Fetching News. Status Code: ", resp.StatusCode)
	}
}

// FetchNews func with 2 parameters (query and page) and return the Result struct
//Notice that the search query is URL encoded through the QueryEscape() method.
func (c *Client) FetchNews(searchType, query, page, param string) (*Results, error) {

	var endpoint = ""

	switch {
	case searchType == "ByCountry":
		// here param is the 'country'
		endpoint = fmt.Sprintf("https://newsapi.org/v2/top-headlines?q=%s&country=%s&apiKey=%s&pageSize=%d&page=%s", url.QueryEscape(query), countries[param], c.key, c.PageSize, page)
	case searchType == "Global":
		// here param is the 'domains'
		endpoint = fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&domains=%s&pageSize=%d&page=%s&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(query), param, c.PageSize, page, c.key)
	default:
		log.Fatal("Search type Must be specified. Allowed values: 'Global', 'ByCountry'")
	}

	resp, err := c.http.Get(endpoint)

	// Handle error from the response
	if err != nil {
		log.Fatal("Error getting a response => ", err)
		return nil, err
	}

	defer resp.Body.Close()

	// Response body is converted to a byte slice, if no errors
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Error reading the response => ", err)
		return nil, err
	}

	// this is for Printing Response Body and Ret Code
	//fmt.Println(string(body))
	//fmt.Printf("Response Status: %s\n", resp.Status)

	// checking ret code, http.StatusOk is a const from http pkg
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}

	res := &Results{}
	// func Unmarshal(data []byte, v interface{}) error
	// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
	// If v is nil or not a pointer, Unmarshal returns an InvalidUnmarshalError.
	return res, json.Unmarshal(body, res)

}
