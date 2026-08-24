package main

import (
	"net/http"
	"time"

	"github.com/its-me-sv/chirpy/internal/auth"
)

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	type responseBody struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	refreshTokenFromDb, err := cfg.db.GetRefreshTokenByToken(req.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to get refresh token", nil)
		return
	}

	if refreshTokenFromDb.RevokedAt.Valid || refreshTokenFromDb.ExpiresAt.Before(time.Now().UTC()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired", nil)
		return
	}

	newToken, err := auth.MakeJWT(refreshTokenFromDb.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create new token", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, responseBody{
		Token: newToken,
	})
}

func (cfg *apiConfig) handleRevokeRefreshToken(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	if err = cfg.db.RevokeRefreshTokenByToken(req.Context(), refreshToken); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to revoke refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
