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

type code struct {
	LongURL string `json:"long_url"`
}

var links = []URL{
	{LongURL: "https://www.google.com/", ShortCode: "000001"},
	{LongURL: "https://www.youtube.com/", ShortCode: "000002"},
}

func getLinks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, links)
}

func importLink(c *gin.Context) {
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

func requestLink(c *gin.Context) {
	code := c.Param("short_code")
	URL, err := getLink(code)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Link not found."})
		return
	}

	c.IndentedJSON(http.StatusOK, URL)
}

func getLink(code string) (*URL, error) {
	for i, l := range links {
		if l.ShortCode == code {
			return &links[i], nil
		}
	}
	return nil, errors.New("code not found")
}

func createLink(c *gin.Context) {
	var url code

	if err := c.BindJSON(&url); err != nil {
		return
	}
	for i, l := range links {
		if l.LongURL == url.LongURL {
			c.IndentedJSON(http.StatusFound, links[i])
			return
		}
	}

	var newLink URL
	newLink.LongURL = url.LongURL
	newLink.ShortCode = uuid.New().String()[:6] // Get first 6 characters from UUID

	links = append(links, newLink)
	c.IndentedJSON(http.StatusCreated, newLink)
}

func main() {
	router := gin.Default()
	router.GET("/links", getLinks)                // curl localhost:8080/links
	router.GET("/links/:short_code", requestLink) // curl localhost:8080/links/000002
	// router.POST("/links", importLink)             // curl localhost:8080/links --include --header "Content-Type: application/json" -d @body.json --request "POST"
	router.POST("/links", createLink)
	router.Run("localhost:8080")
}
