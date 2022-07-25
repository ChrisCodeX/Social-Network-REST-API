package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ChrisCodeX/REST-API-Go/models"
	"github.com/ChrisCodeX/REST-API-Go/repository"
	"github.com/ChrisCodeX/REST-API-Go/server"
	"github.com/segmentio/ksuid"
)

// Items necessary for the registration of a user
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response
type SignUpResponse struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

func SignUpHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request = SignUpRequest{}

		// Decode the request
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Generate a random id
		id, err := ksuid.NewRandom()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert request data in user struct
		var user = models.User{
			Email:    request.Email,
			Password: request.Password,
			Id:       id.String(),
		}

		// Send the struct to be stored in the database
		err = repository.InsertUser(r.Context(), &user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Response to the client
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SignUpResponse{
			Id:    user.Id,
			Email: user.Email,
		})
		// Message on the server
		log.Println("User registered successfully")
	}
}