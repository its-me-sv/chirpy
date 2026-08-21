package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		CleanedBody string `json:"cleaned_body"`
	}

	reqBody := requestBody{}

	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	if len(reqBody.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, responseBody{CleanedBody: getProfanceReplacedString(reqBody.Body)})
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
