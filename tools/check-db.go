package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run check-db.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  registrations - List pending registrations")
		fmt.Println("  users         - List all users")
		fmt.Println("  sessions      - List active sessions")
		fmt.Println("  approve <registration_id> - Approve a registration")
		return
	}

	db, err := sql.Open("sqlite", "data/zfscloud.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	switch os.Args[1] {
	case "registrations":
		rows, err := db.Query(`
			SELECT registration_id, client_id, hostname, username, os, status, requested_at
			FROM registrations
			WHERE status = 'pending'
		`)
		if err != nil {
			log.Fatal("Query failed:", err)
		}
		defer rows.Close()

		fmt.Println("\nPending Registrations:")
		fmt.Println("======================")
		count := 0
		for rows.Next() {
			var regID, clientID, hostname, username, os, status, requestedAt string
			if err := rows.Scan(&regID, &clientID, &hostname, &username, &os, &status, &requestedAt); err != nil {
				log.Fatal("Scan failed:", err)
			}
			count++
			fmt.Printf("\nRegistration ID: %s\n", regID)
			fmt.Printf("Client ID: %s\n", clientID)
			fmt.Printf("Hostname: %s\n", hostname)
			fmt.Printf("Username: %s\n", username)
			fmt.Printf("OS: %s\n", os)
			fmt.Printf("Status: %s\n", status)
			fmt.Printf("Requested: %s\n", requestedAt)
		}
		if count == 0 {
			fmt.Println("No pending registrations found.")
		}

	case "users":
		rows, err := db.Query(`
			SELECT user_id, client_id, hostname, quota_bytes, used_bytes, revoked, created_at
			FROM users
		`)
		if err != nil {
			log.Fatal("Query failed:", err)
		}
		defer rows.Close()

		fmt.Println("\nAll Users:")
		fmt.Println("==========")
		count := 0
		for rows.Next() {
			var userID, clientID, hostname, createdAt string
			var quota, used int64
			var revoked bool
			if err := rows.Scan(&userID, &clientID, &hostname, &quota, &used, &revoked, &createdAt); err != nil {
				log.Fatal("Scan failed:", err)
			}
			count++
			fmt.Printf("\nUser ID: %s\n", userID)
			fmt.Printf("Client ID: %s\n", clientID)
			fmt.Printf("Hostname: %s\n", hostname)
			fmt.Printf("Quota: %d GB\n", quota/(1024*1024*1024))
			fmt.Printf("Used: %d MB\n", used/(1024*1024))
			fmt.Printf("Revoked: %v\n", revoked)
			fmt.Printf("Created: %s\n", createdAt)
		}
		if count == 0 {
			fmt.Println("No users found.")
		}

	case "sessions":
		rows, err := db.Query(`
			SELECT session_id, user_id, last_heartbeat, created_at
			FROM sessions
		`)
		if err != nil {
			log.Fatal("Query failed:", err)
		}
		defer rows.Close()

		fmt.Println("\nActive Sessions:")
		fmt.Println("================")
		count := 0
		for rows.Next() {
			var sessionID, userID, lastHeartbeat, createdAt string
			if err := rows.Scan(&sessionID, &userID, &lastHeartbeat, &createdAt); err != nil {
				log.Fatal("Scan failed:", err)
			}
			count++
			fmt.Printf("\nSession ID: %s\n", sessionID)
			fmt.Printf("User ID: %s\n", userID)
			fmt.Printf("Last Heartbeat: %s\n", lastHeartbeat)
			fmt.Printf("Created: %s\n", createdAt)
		}
		if count == 0 {
			fmt.Println("No active sessions found.")
		}

	case "approve":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run check-db.go approve <registration_id>")
			return
		}
		regID := os.Args[2]

		// Update registration status
		_, err := db.Exec(`
			UPDATE registrations
			SET status = 'approved', processed_at = datetime('now')
			WHERE registration_id = ?
		`, regID)
		if err != nil {
			log.Fatal("Failed to approve registration:", err)
		}

		// Get registration details
		var clientID, hostname, username, publicKey, encryptedPassphrase string
		err = db.QueryRow(`
			SELECT client_id, hostname, username, public_key, encrypted_passphrase
			FROM registrations
			WHERE registration_id = ?
		`, regID).Scan(&clientID, &hostname, &username, &publicKey, &encryptedPassphrase)
		if err != nil {
			log.Fatal("Failed to get registration details:", err)
		}

		// Create user
		userID := "user-" + clientID[7:] // Remove "client-" prefix
		_, err = db.Exec(`
			INSERT INTO users (user_id, client_id, public_key, encrypted_passphrase, zfs_dataset, hostname, quota_bytes, used_bytes, revoked, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, 0, FALSE, datetime('now'))
		`, userID, clientID, publicKey, encryptedPassphrase, "tank/users/"+userID, hostname, 10*1024*1024*1024) // 10GB quota
		if err != nil {
			log.Fatal("Failed to create user:", err)
		}

		fmt.Printf("\n✓ Registration approved!\n")
		fmt.Printf("Registration ID: %s\n", regID)
		fmt.Printf("User ID: %s\n", userID)
		fmt.Printf("Client ID: %s\n", clientID)
		fmt.Printf("Hostname: %s\n", hostname)
		fmt.Printf("Quota: 10 GB\n")

	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}
