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

type userCreds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	reqBody := userCreds{}
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

	type responseBody struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	reqBody := userCreds{}
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

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	reqBody := userCreds{}
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request body", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	userFromDb, err := cfg.db.GetUserByID(req.Context(), userId)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "user not found", nil)
		return
	}

	userFromDbWithEmail, err := cfg.db.GetUserByEmail(req.Context(), reqBody.Email)
	if err == nil && userFromDbWithEmail.ID != userId {
		respondWithError(w, http.StatusUnauthorized, "email already in use", nil)
		return
	}

	newHashedPwd, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update credentials", err)
		return
	}

	updateCredsParams := database.UpdateUserCredsByUserIdParams{
		Email:          reqBody.Email,
		HashedPassword: newHashedPwd,
		ID:             userFromDb.ID,
	}
	newUserData, err := cfg.db.UpdateUserCredsByUserId(req.Context(), updateCredsParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update credentials", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        newUserData.ID,
		CreatedAt: newUserData.CreatedAt,
		UpdatedAt: newUserData.UpdatedAt,
		Email:     newUserData.Email,
	})
}
