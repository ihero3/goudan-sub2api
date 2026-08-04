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

	// List users
	rows, err := db.QueryContext(ctx, "SELECT id, email, role, status FROM users ORDER BY id LIMIT 10")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Users in database:")
	for rows.Next() {
		var id int64
		var email, role, status string
		rows.Scan(&id, &email, &role, &status)
		fmt.Printf("  id=%d, email=%s, role=%s, status=%s\n", id, email, role, status)
	}

	// Find target user
	var userID int64
	var email, role, status string
	err = db.QueryRowContext(ctx, "SELECT id, email, role, status FROM users WHERE email = 'admin@sub2api.local' LIMIT 1").Scan(&userID, &email, &role, &status)
	if err != nil {
		fmt.Printf("admin@sub2api.local not found, using first user...\n")
		err = db.QueryRowContext(ctx, "SELECT id, email, role, status FROM users ORDER BY id LIMIT 1").Scan(&userID, &email, &role, &status)
		if err != nil {
			fmt.Fprintf(os.Stderr, "No users found: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Target user: id=%d, email=%s, role=%s, status=%s\n", userID, email, role, status)

	// Reset password
	newPassword := "TempTest123!"
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
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Password: %s\n", newPassword)
}