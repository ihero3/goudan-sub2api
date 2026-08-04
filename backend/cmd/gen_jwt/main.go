package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := "74f2460691507b5537ee29ca55bfbbea95fb77aff763b3f3a7ffb231a305243a"
	userID := int64(1)
	email := "admin@sub2api.local"
	role := "admin"
	tokenVersion := int64(1)

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id":       userID,
		"email":         email,
		"role":          role,
		"token_version": tokenVersion,
		"exp":           jwt.NewNumericDate(expiresAt).Unix(),
		"iat":           jwt.NewNumericDate(now).Unix(),
		"nbf":           jwt.NewNumericDate(now).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated JWT token:")
	fmt.Println(tokenString)
	fmt.Println()
	fmt.Printf("Expires at: %s\n", expiresAt.Format(time.RFC3339))

	// Also generate a random user_id version
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	hexID := hex.EncodeToString(randBytes)
	fmt.Printf("\nRandom hex ID for reference: %s\n", hexID)
}