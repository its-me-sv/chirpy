package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/its-me-sv/chirpy/internal/auth"
)

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil || apiKey != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type requestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	reqBody := requestBody{}
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if reqBody.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := cfg.db.GetUserByID(req.Context(), reqBody.Data.UserID); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := cfg.db.UpgradeUserToChirpyRed(req.Context(), reqBody.Data.UserID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
