package pokemons

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"example.com/pokemon-handbook/config"
)

type pokemon struct {
	ID          int64  `bson:"_id" json:"id"`
	Name        string `bson:"name" json:"name"`
	IsLegendary bool   `bson:"is_legendary" json:"is_legendary"`
	Color       string `bson:"color" json:"color"`
}

// Post Pokemon godoc
// @title        Post Pokemon
// @summary      Post pokemon to the MongoDB
// @description  Post a pokemon to the MongoDB. If the database doesn't exist, create and insert a new value. Pass values in json format.
// @produce      json
// @success      201 {object} pokemon
// @failure      400 {string} string "object can't be parsed into JSON"
// @failure      409 {string} string "a pokemon with such id already exists"
// @router       /pokemons [post]
func PostPokemon(c *gin.Context) {
	var newPokemon pokemon

	if err := c.BindJSON(&newPokemon); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "object can't be parsed into JSON"})
		return
	}

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.CollectionName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := collection.InsertOne(context.Background(), newPokemon)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "a pokemon with such id already exists"})
		}
		fmt.Println(err)
		return
	}
	id := res.InsertedID
	fmt.Printf("id value %v\n", id)

	// pokemons = append(pokemons, newPokemon)
	c.IndentedJSON(http.StatusCreated, newPokemon)
}

// GetPokemons godoc
// @title        Get Pokemons
// @summary      Retrieves all pokemons from the MongoDB
// @description  Get all pokemons from the MongoDB. Pass values in json format.
// @produce      json
// @success      200 {array} pokemon
// @failure      400 {string} string "object can't be parsed into JSON"
// @failure      404 {string} string "Error: Not Found"
// @router       /pokemons [get]
func GetPokemons(c *gin.Context) {
	var pokemons = []pokemon{}

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.CollectionName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cur, err := collection.Find(context.Background(), bson.D{}, opts)
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

// GetPokemonByID godoc
// @title        Get Pokemon By ID
// @summary      Retrieve pokemon from the MongoDB based on given ID
// @description  Get a pokemon from the MongoDB by ID. Pass values in json format. If there aren't any pokemon with the ID gives a message "pokemon not found".
// @produce      json
// @success      200 {object} pokemon
// @failure      406 {string} string "must be a number"
// @failure      404 {string} string "pokemon not found"
// @router       /pokemons/{id} [get]
func GetPokemonByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"message": "must be a number"})
		return
	}

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.CollectionName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	result := pokemon{}
	err = collection.FindOne(context.Background(), bson.D{{Key: "_id", Value: id}}).Decode(&result)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "pokemon not found"})
		// log.Fatal(err)
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

// UpdatePokemonByID godoc
// @title        Update Pokemon By ID
// @summary      Update pokemon's data in the MongoDB based on given ID
// @description  Update an existing pokemon in the MongoDB by ID. Pass values in json format. If there isn't pokemon with the ID creates a new pokemon.
// @produce      json
// @success      200 {string} string "pokemon was updated"
// @success      201 {object} pokemon
// @failure      406 {string} string "must be a number"
// @failure      400 {string} string "object can't be parsed into JSON"
// @failure      406 {string} string "pokemon's id cannot be changed"
// @router       /pokemons/{id} [put]
func UpdatePokemonByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"message": "must be a number"})
		return
	}
	var newPokemon pokemon

	if err := c.BindJSON(&newPokemon); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "object can't be parsed into JSON"})
		return
	}

	if newPokemon.ID != id {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"message": "pokemon's id cannot be changed"})
		return
	}

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.CollectionName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: newPokemon}}

	result, err := collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		log.Fatal(err)
	}

	if result.MatchedCount != 0 {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "pokemon was updated"})
		return
	}
	if result.UpsertedCount != 0 {
		// pokemons = append(pokemons, newPokemon)
		fmt.Printf("inserted a new pokemon with ID %v\n", result.UpsertedID)
		c.IndentedJSON(http.StatusCreated, newPokemon)
	}
}

// DeletePokemonByID godoc
// @title        Delete Pokemon By ID
// @summary      Delete pokemon in the MongoDB based on given ID
// @description  Delete an existing pokemon in the MongoDB by ID and gives a message. Pass values in json format. If there isn't pokemon with the ID gives a message.
// @produce      json
// @success      200 {object} pokemon "pokemon was deleted"
// @failure      201 {string} string "pokemon not found"
// @router       /pokemons/{id} [delete]
func DeletePokemonByID(c *gin.Context) {
	id := c.Param("id")

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.CollectionName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := collection.DeleteOne(context.Background(), bson.D{{Key: "_id", Value: id}})
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

// DeleteAllPokemons godoc
// @title        Delete All Pokemons
// @summary      Delete all pokemons in the MongoDB
// @description  Delete all existing pokemons in the MongoDB and gives a message "all pokemons are deleted". Pass values in json format. If there aren't pokemons in the database gives a message "pokemons not found".
// @produce      json
// @success      200 {object} pokemon "all pokemons was deleted"
// @failure      404 {string} string "pokemons not found"
// @router       /pokemons [delete]
func DeleteAllPokemons(c *gin.Context) {
	collection, cancel, err := config.ConnectToMongoDB(config.Conf.CollectionName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := collection.DeleteMany(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
		return
	}
	// pokemons = make(map[int]pokemon)
	if res.DeletedCount == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "pokemons not found"})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "all pokemons was deleted"})
	}
}
