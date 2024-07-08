package main

import (
	"net/http"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type URL struct {
	LongURL   string `json:"long_url"`
	ShortCode string `json:"short_code"`
}

var links = []URL{
	{LongURL: "https://www.google.com/", ShortCode: "000001"},
	{LongURL: "https://www.youtube.com/", ShortCode: "000002"},
}

func getLinks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, links)
}

func addLink(c *gin.Context) {
	var newLink URL

	if err := c.BindJSON(&newLink); err != nil {
		return
	}

	for _, link := range links {
		if link == newLink {
			c.IndentedJSON(http.StatusFound, gin.H{"message": "Link already in memory."})
			return
		}
	}

	links = append(links, newLink)
	c.IndentedJSON(http.StatusCreated, newLink)
}

func requestLong(c *gin.Context) {
	code := c.Param("short_code")
	URL, err := getLong(code)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Link not found."})
		return
	}

	c.IndentedJSON(http.StatusOK, URL)
}

func getLong(code string) (*URL, error) {
	for i, l := range links {
		if l.ShortCode == code {
			return &links[i], nil
		}
	}
	return nil, errors.New("code not found")
}

func createLink(c *gin.Context) { //HTTP/1.1 404 Not Found
	var code string
	if err := c.BindJSON(&code); err != nil {
		// Handle error (e.g., log the error, return an error response)
		c.IndentedJSON(http.StatusPreconditionFailed, gin.H{"message": "Could not read input."})
		return
	}
	for i, l := range links {
		if l.LongURL == code {
			c.IndentedJSON(http.StatusFound, links[i])
			return
		}
	}
	var newLink URL
	newLink.LongURL = code
	newLink.ShortCode = uuid.New().String()[:6] // Get first 6 characters from UUID

	links = append(links, newLink)
	c.IndentedJSON(http.StatusCreated, newLink)
}

func main() {
	router := gin.Default()
	router.GET("/links", getLinks)                // curl localhost:8080/links
	router.GET("/links/:short_code", requestLong) // curl localhost:8080/links/000002
	router.POST("/links", addLink)                // curl localhost:8080/links --include --header "Content-Type: application/json" -d @body.json --request "POST"
	router.POST("/links/", createLink)
	router.Run("localhost:8080")
}
