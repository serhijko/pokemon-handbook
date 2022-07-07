package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type pokemon struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsLegendary bool   `json:"is_legendary"`
	Color       string `json:"color"`
}

type Config struct {
	DatabaseURL    string
	DatabaseName   string
	CollectionName string
	URL            string
}

var config Config

func ReadConfig() Config {
	var configFile = "properties.ini"
	_, err := os.Stat(configFile)
	if err != nil {
		log.Fatal("Config file is missing: ", configFile)
	}

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal(err)
	}
	return config
}

func connectToMongoDB() (*mongo.Collection, context.CancelFunc, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.DatabaseURL))
	var collection *mongo.Collection
	if err == nil {
		fmt.Printf("Client value %v\n", client)

		collection = client.Database(config.DatabaseName).Collection(config.CollectionName)
		fmt.Printf("Collection value %v\n", collection)
	}

	return collection, cancel, err
}

func init() {
	fmt.Println("This is init")

	config := ReadConfig()
	fmt.Printf(
		"DatabaseURD: %s\nDatabaseName: %s\nCollectionName: %s\nURL: %s\n",
		config.DatabaseURL, config.DatabaseName, config.CollectionName, config.URL,
	)
}

func main() {
	fmt.Println("This is main")

	router := gin.Default()
	router.POST("/pokemons", postPokemons)
	router.GET("/pokemons", getPokemons)
	router.GET("/pokemons/:id", getPokemonByID)
	router.PUT("/pokemons/:id", updatePokemonByID)
	router.DELETE("/pokemons/:id", deletePokemonByID)

	router.Run(config.URL)
}

func postPokemons(c *gin.Context) {
	var newPokemon pokemon

	if err := c.BindJSON(&newPokemon); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "object can't be parsed into JSON"})
		return
	}

	collection, cancel, err := connectToMongoDB()
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := collection.InsertOne(context.Background(), newPokemon)
	if err != nil {
		fmt.Println(err)
		return
	}
	id := res.InsertedID
	fmt.Printf("id value %v\n", id)

	// pokemons = append(pokemons, newPokemon)
	c.IndentedJSON(http.StatusCreated, newPokemon)
}

func getPokemons(c *gin.Context) {
	var pokemons = []pokemon{}

	collection, cancel, err := connectToMongoDB()
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		// To decode into a struct, use cursor.Decode()
		result := pokemon{}
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Pokemon entry: %v\n", result)
		pokemons = append(pokemons, result)

		// To get the raw bson bytes use cursor.Current
		raw := cur.Current
		fmt.Printf("Raw result entry: %v\n", raw)
	}
	if err := cur.Err(); err != nil {
		fmt.Println(err)
		return
	}
	c.IndentedJSON(http.StatusOK, pokemons)
}

func getPokemonByID(c *gin.Context) {
	id := c.Param("id")

	collection, cancel, err := connectToMongoDB()
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	result := pokemon{}
	err = collection.FindOne(context.Background(), bson.D{{Key: "id", Value: id}}).Decode(&result)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "pokemon not found"})
		// log.Fatal(err)
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

func updatePokemonByID(c *gin.Context) {
	var newPokemon pokemon

	if err := c.BindJSON(&newPokemon); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "object can't be parsed into JSON"})
		return
	}

	collection, cancel, err := connectToMongoDB()
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	opts := options.FindOneAndUpdate().SetUpsert(false)
	filter := bson.D{{Key: "id", Value: newPokemon.ID}}
	update := bson.D{{Key: "$set", Value: newPokemon}}
	var updatedPokemon bson.M
	err = collection.FindOneAndUpdate(
		context.Background(),
		filter,
		update,
		opts,
	).Decode(&updatedPokemon)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// pokemons = append(pokemons, newPokemon)
			res, err := collection.InsertOne(context.Background(), newPokemon)
			if err != nil {
				fmt.Println(err)
				return
			}
			id := res.InsertedID
			fmt.Printf("id value %v\n", id)

			c.IndentedJSON(http.StatusCreated, newPokemon)
			return
		}
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "pokemon was updated"})
}

func deletePokemonByID(c *gin.Context) {
	id := c.Param("id")

	collection, cancel, err := connectToMongoDB()
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := collection.DeleteOne(context.Background(), bson.D{{Key: "id", Value: id}})
	if err != nil {
		log.Fatal(err)
		return
	}
	// pokemons = append(pokemons[:s], pokemons[s+1:]...)
	if res.DeletedCount == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "pokemon not found"})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "pokemon was deleted"})
	}
}
