package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := "74f2460691507b5537ee29ca55bfbbea95fb77aff763b3f3a7ffb231a305243a"
	tokenString := os.Args[1]

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "Invalid token format\n")
		os.Exit(1)
	}

	headerJSON, _ := base64.RawURLEncoding.DecodeString(parts[0])
	var header map[string]interface{}
	json.Unmarshal(headerJSON, &header)
	fmt.Printf("Header: %v\n", header)

	payloadJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var payload map[string]interface{}
	json.Unmarshal(payloadJSON, &payload)
	fmt.Printf("Payload: %v\n", payload)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		if errors.Is(err, jwt.ErrTokenExpired) {
			fmt.Println("Token is expired")
		}
	} else if token.Valid {
		fmt.Println("Token is VALID")
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			fmt.Printf("Claims: %v\n", claims)
		}
	} else {
		fmt.Println("Token is INVALID")
	}
}