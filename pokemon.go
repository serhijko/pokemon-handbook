package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type pokemon struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsLegendary bool   `json:"is_legendary"`
	Color       string `json:"color"`
}

var pokemons = []pokemon{
	{ID: "1", Name: "bulbasaur", IsLegendary: false, Color: "green"},
	{ID: "4", Name: "charmander", IsLegendary: false, Color: "red"},
	{ID: "25", Name: "pikachu", IsLegendary: false, Color: "yellow"},
	{ID: "54", Name: "psyduck", IsLegendary: false, Color: "yellow"},
}

func main() {
	router := gin.Default()
	router.POST("/pokemons", postPokemons)
	router.GET("/pokemons", getPokemons)
	router.GET("/pokemons/:id", getPokemonByID)
	router.PUT("/pokemons/:id", updatePokemonByID)
	router.DELETE("/pokemons/:id", deletePokemonByID)

	router.Run("localhost:8080")
}

func postPokemons(c *gin.Context) {
	var newPokemon pokemon

	if err := c.BindJSON(&newPokemon); err != nil {
		return
	}

	pokemons = append(pokemons, newPokemon)
	c.IndentedJSON(http.StatusCreated, newPokemon)
}

func getPokemons(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, pokemons)
}

func getPokemonByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range pokemons {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "pokemon not found"})
}

func updatePokemonByID(c *gin.Context) {
	var newPokemon pokemon

	if err := c.BindJSON(&newPokemon); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "object can't be parsed into JSON"})
	}

	for s, a := range pokemons {
		if a.ID == newPokemon.ID {
			pokemons[s] = newPokemon
			c.IndentedJSON(http.StatusOK, gin.H{"message": "pokemon was updated"})
			return
		}
	}

	pokemons = append(pokemons, newPokemon)
	c.IndentedJSON(http.StatusCreated, newPokemon)
}

func deletePokemonByID(c *gin.Context) {
	id := c.Param("id")

	for s, a := range pokemons {
		if a.ID == id {
			pokemons = append(pokemons[:s], pokemons[s+1:]...)
			c.IndentedJSON(http.StatusOK, gin.H{"message": "pokemon was deleted"})
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "pokemon not found"})
}
