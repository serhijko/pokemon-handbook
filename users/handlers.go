package users

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"example.com/pokemon-handbook/config"
)

type user struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func CheckAdminInDB() {
	newUser := user{
		Login:    config.Conf.UserName,
		Password: config.Conf.Password,
		Role:     "admin",
	}

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.UserCollecName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	var existedAdminUser bson.M
	err = collection.FindOne(
		context.Background(),
		bson.D{{Key: "role", Value: "admin"}},
	).Decode(&existedAdminUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			res, err := collection.InsertOne(context.Background(), newUser)
			if err != nil {
				fmt.Println(err)
				return
			}
			id := res.InsertedID
			fmt.Printf("id value %v\n", id)

			// users = append(users, newUser)
			fmt.Println("The user with admin role is added to the users collection of the database.")
			return
		}
		log.Fatal(err)
	}
	fmt.Printf("A user with admin role already exists: %v\n", existedAdminUser)
}

// Post User godoc
// @title        Post User
// @summary      Post user to the MongoDB
// @description  Post a user to the MongoDB. If the database doesn't exist, create and insert a new value. Pass values in json format.
// @produce      json
// @success      201 {object} user
// @failure      400 {string} string "object can't be parsed into JSON"
// @router       /users [post]
func PostUser(c *gin.Context) {
	var newUser user

	if err := c.BindJSON(&newUser); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "object can't be parsed into JSON"})
		return
	}

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.UserCollecName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = collection.Find(context.Background(), bson.D{{Key: "login", Value: newUser.Login}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			res, err := collection.InsertOne(context.Background(), newUser)
			if err != nil {
				fmt.Println(err)
				return
			}
			id := res.InsertedID
			fmt.Printf("id value %v\n", id)

			// users = append(users, newUser)
			c.IndentedJSON(http.StatusCreated, newUser)
			return
		}
		log.Fatal(err)
	}
	fmt.Printf("A user with such login already exists, choose another login!")
}

// GetUsers godoc
// @title        Get Users
// @summary      Retrieves all users from the MongoDB
// @description  Get all users from the MongoDB. Pass values in json format.
// @produce      json
// @success      200 {array} user
// @failure      400 {string} string "object can't be parsed into JSON"
// @failure      404 {string} string "Error: Not Found"
// @router       /users [get]
func GetUsers(c *gin.Context) {
	var users = []user{}

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.UserCollecName)
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
		result := user{}
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User entry: %v\n", result)
		users = append(users, result)

		// To get the raw bson bytes use cursor.Current
		raw := cur.Current
		fmt.Printf("Raw result entry: %v\n", raw)
	}
	if err := cur.Err(); err != nil {
		fmt.Println(err)
		return
	}
	c.IndentedJSON(http.StatusOK, users)
}

// GetUserByID godoc
// @title        Get User By Login
// @summary      Retrieve user from the MongoDB based on given Login
// @description  Get a user from the MongoDB by given login. Pass values in json format. If there aren't any users with the login gives a message "user not found".
// @produce      json
// @success      200 {object} user
// @failure      404 {string} string "user not found"
// @router       /users/{id} [get]
func GetUserByLogin(c *gin.Context) {
	login := c.Param("id")

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.UserCollecName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	result := user{}
	err = collection.FindOne(context.Background(), bson.D{{Key: "login", Value: login}}).Decode(&result)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		// log.Fatal(err)
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

// UpdateUserByID godoc
// @title        Update User By ID
// @summary      Update user's data in the MongoDB based on given ID
// @description  Update an existing user in the MongoDB by ID. Pass values in json format. If there isn't user with the ID creates a new user.
// @produce      json
// @success      200 {object} user
// @failure      201      {string}  string        ""
// @router       /users/{id} [put]
func UpdateUserByLogin(c *gin.Context) {
	login := c.Param("id")
	var newUser user

	if err := c.BindJSON(&newUser); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "object can't be parsed into JSON"})
		return
	}

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.UserCollecName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{Key: "login", Value: login}}
	update := bson.D{{Key: "$set", Value: newUser}}
	var updatedPokemon bson.M
	err = collection.FindOneAndUpdate(
		context.Background(),
		filter,
		update,
		opts,
	).Decode(&updatedPokemon)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// users = append(users, newUser)
			fmt.Printf("id value %v\n", newUser.Login)
			c.IndentedJSON(http.StatusCreated, newUser)
			return
		}
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "user was updated"})
}

// DeleteUserByLogin godoc
// @title        Delete User By Login
// @summary      Delete user in the MongoDB based on given login
// @description  Delete an existing user in the MongoDB by login and gives a message. Pass values in json format. If there isn't user with the login gives a message.
// @produce      json
// @success      200 {object} user "user was deleted"
// @failure      201 {string} string "user not found"
// @router       /users/{id} [delete]
func DeleteUserByLogin(c *gin.Context) {
	login := c.Param("id")

	collection, cancel, err := config.ConnectToMongoDB(config.Conf.UserCollecName)
	defer cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := collection.DeleteOne(context.Background(), bson.D{{Key: "login", Value: login}})
	if err != nil {
		log.Fatal(err)
		return
	}
	// users = append(users[:s], users[s+1:]...)
	if res.DeletedCount == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "user was deleted"})
	}
}
