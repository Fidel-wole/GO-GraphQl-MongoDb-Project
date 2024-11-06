package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Fidel-wole/gql/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var connectionString = ""


type DB struct {
	client *mongo.Client
}

// Connect establishes a connection to the MongoDB server and returns a client instance wrapped in DB.
func Connect() *DB {
	// Create a context with timeout for the connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize the client with the connection string and connect
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")

	return &DB{client: client}
}


func (db *DB) GetJob(id string) (*model.JobListing, error) {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert the ID from a string to a MongoDB ObjectID
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// Return the error to the caller instead of just printing it
		return nil, fmt.Errorf("invalid ID format: %v", err)
	}
	filter := bson.M{"_id": _id}

	var jobListing model.JobListing
	err = jobCollec.FindOne(ctx, filter).Decode(&jobListing)
	if err != nil {
		// Return the error instead of just printing it
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("job listing not found: %v", err)
		}
		return nil, fmt.Errorf("error finding job listing: %v", err)
	}

	return &jobListing, nil
}

func (db *DB) GetJobs() ([]*model.JobListing, error) {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.D{}
	cursor, err := jobCollec.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobListings []*model.JobListing
	for cursor.Next(ctx) {
		var jobListing model.JobListing
		if err := cursor.Decode(&jobListing); err != nil {
			return nil, err
		}

		jobListings = append(jobListings, &jobListing)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return jobListings, nil
}



func (db *DB) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jobListing := model.JobListing{
		Title:       jobInfo.Title,
		Description: jobInfo.Description,
		Company:     jobInfo.Company,
		URL:         jobInfo.URL,
	}

	res, err := jobCollec.InsertOne(ctx, jobListing)
	if err != nil {
		fmt.Println("Error creating job listing:", err)
		return nil
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		jobListing.ID = oid.Hex() 
	} else {
		fmt.Println("Error converting InsertedID to ObjectID")
		return nil
	}

	return &jobListing
}

func (db *DB) UpdateJobListing(jobId string, jobInfo model.UpdateJobListingInput) *model.JobListing {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		fmt.Println("Invalid job ID:", err)
		return nil
	}

	update := bson.M{"$set": jobInfo}

	// Update the document
	var updatedJobListing model.JobListing
	err = jobCollec.FindOneAndUpdate(ctx, bson.M{"_id": _id}, update).Decode(&updatedJobListing)
	if err != nil {
		fmt.Println("Error updating job listing:", err)
		return nil
	}

	return &updatedJobListing
}

func (db *DB) DeleteJobListing(jobId string) *model.DeleteJobResponse {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		fmt.Println("Invalid job ID:", err)
	}

	_, err = jobCollec.DeleteOne(ctx, bson.M{"_id": _id})
	if err != nil {
		fmt.Println("Error deleting job listing:", err)
	}

	return &model.DeleteJobResponse{DeleteJobID: jobId}
}
