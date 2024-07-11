package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		connectAndCreate(newLink)
	}
}

func connectAndCreate(newLink URL) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	URI := "mongodb+srv://pvasilyev:ccCTkS1UnwiAuxr4@apiproject0.uugqh4j.mongodb.net/?retryWrites=true&w=majority&appName=APIProject0"
	opts := options.Client().ApplyURI(URI).SetServerAPIOptions(serverAPI)

	ctx := context.Background()

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	// Get the database object
	db := client.Database("URL_Shortener_Database")

	// Get the collection object (or create it if it doesn't exist)
	col := db.Collection("URLs")

	_, err = col.InsertOne(ctx, newLink)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("URL mapping saved!")
}

func main() {
	router := gin.Default()
	router.GET("/links", getLinks)                // curl localhost:8080/links
	router.GET("/links/:short_code", requestLink) // curl localhost:8080/links/000002
	router.POST("/links", createLink)             // curl localhost:8080/links --include --header "Content-Type: application/json" -d '{"long_url": "https://gmail.com/"}' --request "POST"
	router.Run("localhost:8080")
}
