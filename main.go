package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tasks/controladores"
)

var (
	mongoURI = ""
	port     = ""
)

func main() {

	// varible para congigurar la variable de entorno de la Bd
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal("Failed to disconnect from MongoDB:", err)
		}
	}()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	r := mux.NewRouter()
	setupRoutes(r, client)

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	if err := http.ListenAndServe(":"+port, handlers.CORS(originsOk, headersOk, methodsOk)(r)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRoutes(r *mux.Router, client *mongo.Client) {
	r.HandleFunc("/tasks", makeTaskHandler(client, controladores.GetTasks)).Methods("GET")
	r.HandleFunc("/tasks", makeTaskHandler(client, controladores.CreateTask)).Methods("POST")
	r.HandleFunc("/tasks/{id}", makeTaskHandler(client, controladores.UpdateTask)).Methods("PUT")
	r.HandleFunc("/tasks/{id}", makeTaskHandler(client, controladores.DeleteTask)).Methods("DELETE")
}

func makeTaskHandler(client *mongo.Client, handler func(*mongo.Client, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in handler: %v", r)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		handler(client, w, r)
	}
}
