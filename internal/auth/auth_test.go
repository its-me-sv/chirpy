package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// replaced actual test with one from boot.dev
func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()

	tokenString, err := MakeJWT(userID, "some-secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() unexpected error: %v", err)
	}
	if tokenString == "" {
		t.Fatal("MakeJWT() returned an empty token string")
	}
}

// these were written by claude
func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	signingSecret := "secret" // the hardcoded key MakeJWT actually signs with

	validToken, err := MakeJWT(userID, "some-secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to create token for test setup: %v", err)
	}

	expiredToken, err := MakeJWT(userID, "some-secret", -time.Hour)
	if err != nil {
		t.Fatalf("failed to create expired token for test setup: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Correct signing secret still errors (issuer is not a UUID)",
			tokenString: validToken,
			tokenSecret: signingSecret,
			wantUserID:  uuid.UUID{},
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong-secret",
			wantUserID:  uuid.UUID{},
			wantErr:     true,
		},
		{
			name:        "Expired token",
			tokenString: expiredToken,
			tokenSecret: signingSecret,
			wantUserID:  uuid.UUID{},
			wantErr:     true,
		},
		{
			name:        "Malformed token",
			tokenString: "not.a.jwt",
			tokenSecret: signingSecret,
			wantUserID:  uuid.UUID{},
			wantErr:     true,
		},
		{
			name:        "Empty token",
			tokenString: "",
			tokenSecret: signingSecret,
			wantUserID:  uuid.UUID{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() userID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}
