package responseusage
package main

import (
	"log"
	"net/http"

	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
)

// Example: Using the response package in a simple handler

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	r := chi.NewRouter()

	// Success example
	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "id")
		
		if userID == "" {
			response.ValidationError(w, "id", "User ID is required")
			return
		}

		// Simulate getting user from database
		user := User{
			ID:    userID,
			Name:  "John Doe",
			Email: "john@example.com",
		}

		response.Success(w, user)
	})

	// Created example
	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		// Simulate creating user
		newUser := User{
			ID:    "123",
			Name:  "Jane Doe",
			Email: "jane@example.com",
		}

		location := "/api/v1/users/123"
		response.Created(w, newUser, location)
	})

	// List with pagination example
	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		users := []User{
			{ID: "1", Name: "User 1", Email: "user1@example.com"},
			{ID: "2", Name: "User 2", Email: "user2@example.com"},
		}

		pagination := &response.PaginationInfo{
			Total:      100,
			Page:       1,
			PageSize:   10,
			TotalPages: 10,
		}

		response.List(w, users, pagination)
	})

	// Error example
	r.Get("/users/{id}/profile", func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "id")
		
		// Simulate user not found
		if userID == "999" {
			response.Error(w, http.StatusNotFound, response.ErrCodeNotFound, "User not found")
			return
		}

		// Return profile...
		response.Success(w, map[string]string{"profile": "data"})
	})

	// Delete example
	r.Delete("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "id")
		
		// Simulate deletion
		log.Printf("Deleting user: %s", userID)
		
		response.NoContent(w)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
