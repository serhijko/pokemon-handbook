package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	DatabaseURL    string
	DatabaseName   string
	CollectionName string
	UserCollecName string
	URL            string
	UserName       string
	Password       string
	UserName1      string
	Password1      string
}

var Conf Config

func ReadConfig() Config {
	var configFile = "./config/properties.ini"
	_, err := os.Stat(configFile)
	if err != nil {
		log.Fatal("Config file is missing: ", configFile)
	}

	if _, err := toml.DecodeFile(configFile, &Conf); err != nil {
		log.Fatal(err)
	}
	return Conf
}

func ConnectToMongoDB(collectionName string) (*mongo.Collection, context.CancelFunc, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(Conf.DatabaseURL))
	var collection *mongo.Collection
	if err == nil {
		fmt.Printf("Client value %v\n", client)

		collection = client.Database(Conf.DatabaseName).Collection(collectionName)
		fmt.Printf("Collection value %v\n", collection)
	}

	return collection, cancel, err
}
