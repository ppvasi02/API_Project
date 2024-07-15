package main

import (
	"context"
	"fmt"
	"net/http"

	//"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var URI string = "mongodb+srv://pvasilyev:ccCTkS1UnwiAuxr4@apiproject0.uugqh4j.mongodb.net/?retryWrites=true&w=majority&appName=APIProject0"
var ctx context.Context
var db *mongo.Database
var col *mongo.Collection
var client *mongo.Client

var links []URL

func connectToDB() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(URI).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	ctx = context.Background()

	db = client.Database("URL_Shortener_Database")

	col = db.Collection("URLs")
}

type URL struct {
	LongURL   string `bson:"long_url"`
	ShortCode string `bson:"short_code"`
}

type code struct {
	LongURL string `bson:"long_url"`
}

func fillLinkList() error {
	cursor, err := col.Find(ctx, bson.M{}) // Find all documents with an empty filter
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var link URL // Define a Link object for each document
		err := cursor.Decode(&link)
		if err != nil {
			return err
		}
		links = append(links, link)
	}

	if err := cursor.Err(); err != nil {
		return err
	}
	return err
}

func printLinks(c *gin.Context) {
	fillLinkList()
	c.IndentedJSON(http.StatusOK, links)
	links = links[:0]
}

/*func requestLink(c *gin.Context) {
	code := c.Param("short_code")
	URL, err := getLink(code)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Link not found."})
		return
	}

	c.IndentedJSON(http.StatusOK, URL)
}*/

/*func getLink(code string) (*URL, error) {
	for i, l := range links {
		if l.ShortCode == code {
			return &links[i], nil
		}
	}
	return nil, errors.New("code not found")
}*/

func createLink(c *gin.Context) {
	var codeList []code
	var newLink URL

	if err := c.BindJSON(&codeList); err != nil {
		return
	}

	cursor, err := col.Find(ctx, bson.M{}) // Find all documents with an empty filter
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	fillLinkList()

	for cursor.Next(context.Background()) {
		var link URL // Define a Link object for each document
		err := cursor.Decode(&link)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		links = append(links, link)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

		//documents = append(documents, doc)
		links = append(links, newLink)
		c.IndentedJSON(http.StatusCreated, newLink)
	}

	var documents []interface{}
	for _, link := range links {
		doc := bson.D{
			{"long_url", link.LongURL},
			{"short_code", link.ShortCode},
		}
		documents = append(documents, doc)
	}

	// Insert multiple documents in a single operation
	_, err = col.InsertMany(context.Background(), documents)
	if err != nil {
		return
	}

	fmt.Println("URL mappings saved!")
	links = links[:0]
}

func main() {
	connectToDB()
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	router := gin.Default()
	router.GET("/links", printLinks) // curl localhost:8080/links
	//router.GET("/links/:short_code", requestLink) // curl localhost:8080/links/000002
	router.POST("/links", createLink) // curl localhost:8080/links --include --header "Content-Type: application/json" -d '{"long_url": "https://gmail.com/"}' --request "POST"
	router.Run("localhost:8080")
}
