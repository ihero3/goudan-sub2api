package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	connStr := "host=localhost port=5432 user=sub2api password=sub2api dbname=sub2api sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()

	// Parse email and password from args
	email := "admin@example.com"
	newPassword := "admin123"
	if len(os.Args) > 1 {
		email = os.Args[1]
	}
	if len(os.Args) > 2 {
		newPassword = os.Args[2]
	}

	// Find target user by email, or fallback to first user
	var userID int64
	var userEmail, role, status string
	err = db.QueryRowContext(ctx, "SELECT id, email, role, status FROM users WHERE email = $1 LIMIT 1", email).Scan(&userID, &userEmail, &role, &status)
	if err != nil {
		fmt.Printf("%s not found, using first user...\n", email)
		err = db.QueryRowContext(ctx, "SELECT id, email, role, status FROM users ORDER BY id LIMIT 1").Scan(&userID, &userEmail, &role, &status)
		if err != nil {
			fmt.Fprintf(os.Stderr, "No users found: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Target user: id=%d, email=%s, role=%s, status=%s\n", userID, userEmail, role, status)

	// Reset password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to hash: %v\n", err)
		os.Exit(1)
	}

	_, err = db.ExecContext(ctx, "UPDATE users SET password_hash = $1, status = 'active' WHERE id = $2", string(hash), userID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to update: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SUCCESS\n")
	fmt.Printf("Email: %s\n", userEmail)
	fmt.Printf("Password: %s\n", newPassword)
}
