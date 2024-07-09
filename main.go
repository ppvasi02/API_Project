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

var links = []URL{}

func getLinks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, links)
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
	var codeList []code
	// var url code
	var newLink URL

	if err := c.BindJSON(&codeList); err != nil {
		return
	}
outerLoop:
	for i := range codeList {
		for j := range links {
			if links[j].LongURL == codeList[i].LongURL {
				c.IndentedJSON(http.StatusFound, gin.H{"message": "Link already in memory."})
				c.IndentedJSON(http.StatusFound, links[j])
				continue outerLoop // move to next iteration of outerloop
			}
		}
		newLink.LongURL = codeList[i].LongURL
		newLink.ShortCode = uuid.New().String()[:6] // Get first 6 characters from UUID

		links = append(links, newLink)
		c.IndentedJSON(http.StatusCreated, newLink)
	}
}

func main() {
	router := gin.Default()
	router.GET("/links", getLinks)                // curl localhost:8080/links
	router.GET("/links/:short_code", requestLink) // curl localhost:8080/links/000002
	router.POST("/links", createLink)             // curl localhost:8080/links --include --header "Content-Type: application/json" -d '{"long_url": "https://gmail.com/"}' --request "POST"
	router.Run("localhost:8080")
}
