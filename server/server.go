package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/ChrisCodeX/REST-API-Go/database"
	"github.com/ChrisCodeX/REST-API-Go/repository"
	"github.com/ChrisCodeX/REST-API-Go/websocket"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Items that the server need to connect
type Config struct {
	Port        string // Port where it is executed
	JWTSecret   string // Secret key used to generate Tokens
	DatabaseUrl string // Database connection
}

// Interface to be considered a server
type Server interface {
	Config() *Config
	Hub() *websocket.Hub
}

// Element that will handle the server
type Broker struct {
	config *Config
	router *mux.Router    // It defines the API route
	hub    *websocket.Hub // Hub for websocket
}

/* Method that makes the broker a server interface */
// Server Configuration
func (b *Broker) Config() *Config {
	return b.config
}

// Websocket Hub
func (b *Broker) Hub() *websocket.Hub {
	return b.hub
}

// Constructor of Server
func NewServer(ctx context.Context, config *Config) (*Broker, error) {
	// Validations
	if config.Port == "" {
		return nil, errors.New("port is required")
	}
	if config.JWTSecret == "" {
		return nil, errors.New("secret is required")
	}
	if config.DatabaseUrl == "" {
		return nil, errors.New("database url is required")
	}

	broker := &Broker{
		config: config,
		router: mux.NewRouter(),
		hub:    websocket.NewHub(), // Create Hub for upgrade to WebSocket
	}

	return broker, nil
}

// Method that makes the server (Broker) able to start
func (b *Broker) Start(binder func(s Server, r *mux.Router)) {
	// Start the binder
	b.router = mux.NewRouter()

	binder(b, b.router)

	// Cors Handler
	handler := cors.Default().Handler(b.router)
	// Other way
	// handler := cors.AllowAll().Handler(b.router)

	// Assign database
	repo, err := database.NewPostgresRepository(b.config.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	repository.SetRepository(repo)

	// Start WebSocket Connection
	go b.hub.Run()

	// Server started logs
	log.Println("Server started on port", b.Config().Port)
	if err := http.ListenAndServe(b.config.Port, handler); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
