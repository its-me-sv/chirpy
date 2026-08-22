package main

import (
	"errors"
	"net/http"
)

func (cfg *apiConfig) handleReset(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "cannot reset in non-dev enviroments", errors.New("Cannot perform reset in non-dev enviroments"))
		return
	}

	cfg.fileserverHits.Store(0)

	if err := cfg.db.DeleteAllUsers(req.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to delete all users", err)
	}

	w.WriteHeader(http.StatusOK)
}
