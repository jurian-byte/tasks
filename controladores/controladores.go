package controladores

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Task struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	State       bool               `json:"state" bson:"state"`
	Priority    bool               `json:"priority" bson:"priority"`
	CreatedAt   string             `json:"created_at" bson:"created_at"`
}

var client *mongo.Client

func GetTasks(client *mongo.Client, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tasks []Task
	collection := client.Database("jurianlt").Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var task Task
		cursor.Decode(&task)
		tasks = append(tasks, task)
	}
	if err := cursor.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tasks)
}

func CreateTask(client *mongo.Client, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task.ID = primitive.NewObjectID()
	task.CreatedAt = time.Now().Format(time.RFC3339)

	collection := client.Database("jurianlt").Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func UpdateTask(client *mongo.Client, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	taskID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := client.Database("jurianlt").Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"state": task.State,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedTask Task
	err = collection.FindOneAndUpdate(ctx, bson.M{"_id": taskID}, update, opts).Decode(&updatedTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedTask)
}

func DeleteTask(client *mongo.Client, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	taskID, _ := primitive.ObjectIDFromHex(params["id"])
	collection := client.Database("jurianlt").Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := collection.DeleteOne(ctx, bson.M{"_id": taskID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}
