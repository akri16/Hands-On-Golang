package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Person struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname string `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

var client *mongo.Client

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(databases)

	testDb := client.Database("test-db")
	testCollection := testDb.Collection("test-collection")
	awsCollection := testDb.Collection("aws-collection")

	testResult, err := testCollection.InsertOne(ctx, bson.D {
		{Key: "title", Value: "The Polyglot Dev Podcast"},
		{Key: "author", Value: "Amit"},
		{"tags", bson.A{"development", "programming", "coding"}},
	})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(testResult.InsertedID)

	awsResult, err := awsCollection.InsertMany(ctx, []interface{}{
		bson.D{
			{"podcast", testResult.InsertedID},
			{"title", "Episode #1"},
			{"description", "This is the first episode"},
			{"duration", 25},
		},
		bson.D {
			{"podcast", testResult.InsertedID},
			{"title", "Episode #2"},
			{"description", "This is the first episode"},
			{"duration", 55},

		},
	})

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(awsResult.InsertedIDs)
	}

	defer client.Disconnect(ctx)
}