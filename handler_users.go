package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/its-me-sv/chirpy/internal/auth"
	"github.com/its-me-sv/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	reqBody := requestBody{}
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request body", err)
		return
	}

	hashedPwd, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create user", err)
		return
	}

	createUserParams := database.CreateUserParams{
		Email:          reqBody.Email,
		HashedPassword: hashedPwd,
	}
	newUser, err := cfg.db.CreateUser(req.Context(), createUserParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	})
}

func (cfg *apiConfig) handleUserLogin(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type responseBody struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	reqBody := requestBody{}
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request body", err)
		return
	}

	userFromDb, err := cfg.db.GetUserByEmail(req.Context(), reqBody.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User not found", err)
		return
	}

	if hashMatched, err := auth.CheckPasswordHash(reqBody.Password, userFromDb.HashedPassword); err != nil || !hashMatched {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	token, err := auth.MakeJWT(userFromDb.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err)
		return
	}

	saveRefreshTokenParams := database.SaveRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    userFromDb.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	}
	refreshTokenFromDb, err := cfg.db.SaveRefreshToken(req.Context(), saveRefreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, responseBody{
		User: User{
			ID:        userFromDb.ID,
			CreatedAt: userFromDb.CreatedAt,
			UpdatedAt: userFromDb.UpdatedAt,
			Email:     userFromDb.Email,
		},
		Token:        token,
		RefreshToken: refreshTokenFromDb.Token,
	})
}
