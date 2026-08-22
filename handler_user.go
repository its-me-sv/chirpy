package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	type requestBody struct {
		Email string `json:"email"`
	}
	type responseBody struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	reqBody := requestBody{}
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	newUser, err := cfg.db.CreateUser(req.Context(), reqBody.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, responseBody(newUser))
}
