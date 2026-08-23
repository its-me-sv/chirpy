package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/its-me-sv/chirpy/internal/auth"
	"github.com/its-me-sv/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	tokenUserID, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	type requestBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		Chirp
	}

	reqBody := requestBody{}
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request body", err)
		return
	}

	if len(reqBody.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	createChirpParams := database.CreateChirpParams{
		Body:   getProfanceReplacedString(reqBody.Body),
		UserID: tokenUserID,
	}
	dbChirp, err := cfg.db.CreateChirp(req.Context(), createChirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, responseBody{Chirp: Chirp(dbChirp)})
}

var profanceWords = [3]string{"kerfuffle", "sharbert", "fornax"}

func getProfanceReplacedString(orginal string) string {
	words := strings.Split(orginal, " ")

	for i, word := range words {
		lowered := strings.ToLower(word)
		for _, profaneWord := range profanceWords {
			if strings.Contains(lowered, profaneWord) {
				words[i] = strings.ReplaceAll(lowered, profaneWord, "****")
				break
			}
		}
	}

	return strings.Join(words, " ")
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	dbChirps, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to retrieve all chirps", err)
		return
	}

	chirps := make([]Chirp, len(dbChirps))
	for i, chirp := range dbChirps {
		chirps[i] = Chirp(chirp)
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	chirpIDFromPath := req.PathValue("chirpID")
	if chirpIDFromPath == "" {
		respondWithError(w, http.StatusBadRequest, "missing \"chirpID\"", errors.New("missing required param \"chirpID\""))
		return
	}

	chirpID, err := uuid.Parse(chirpIDFromPath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid \"chirpID\"", errors.New("malformed \"chirpID\""))
		return
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp(chirp))
}
