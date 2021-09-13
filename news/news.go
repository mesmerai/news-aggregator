package news

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Client struct {
	http     *http.Client
	key      string
	pageSize int
}

func NewClient(httpClient *http.Client, key string, pageSize int) *Client {
	if pageSize > 100 {
		pageSize = 100
	}
	return &Client{httpClient, key, pageSize}

}

func (c *Client) FetchNews(query, page string) {
	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%s&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(query), c.pageSize, page, c.key)
	//req, err := c.http.NewRequest("GET", restEndpoint, nil)
	resp, err := c.http.Get(endpoint)

	// Handle error from the response
	if err != nil {
		log.Fatal("Error getting a response => ", err)
	}

	defer resp.Body.Close()

	// Read the response / get 'body' in Bytes
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Error reading the response => ", err)
	}

	// for now print Response Body and RetCode
	fmt.Println(string(body))
	fmt.Printf("Response Status: %s\n", resp.Status)

}
