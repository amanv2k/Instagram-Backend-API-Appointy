package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Users struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"Name,omitempty" bson:"Name,omitempty"`
	Email    string             `json:"Email,omitempty" bson:"Email,omitempty"`
	Password string             `json:"Password,omitempty" bson:"Password,omitempty"`
}

var client *mongo.Client

type MyMux struct {
}

func (p *MyMux) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/" {
		// sayhelloName(w, r)
		fmt.Println("mux active")
		return
	} else if request.URL.Path == "/users" {
		fmt.Printf("users mux")
		CreateUserEndpoint(response, request)
		return
	}
	http.NotFound(response, request)
	return
}

func CreateUserEndpoint(response http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		response.Header().Add("content-type", "application/json")
		var users Users
		json.NewDecoder(request.Body).Decode((&users))
		collection := client.Database("testdb").Collection("testcollection")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		result, _ := collection.InsertOne(ctx, users)
		json.NewEncoder(response).Encode(result)
	} else {
		fmt.Printf("Method not allowed")
		http.Redirect(response, request, "/", http.StatusFound)
	}

}

func main() {
	fmt.Println("Starting the app...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	router := &MyMux{}
	http.HandleFunc("/users", CreateUserEndpoint)
	http.ListenAndServe(":12345", router)

}
