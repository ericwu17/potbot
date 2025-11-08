// `plant.go` contains all the endpoints that a plant interacts with
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func handleVerifyPlantCreds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !verifyPlantCreds(w, r) {
		return
	}
	// If we get here, credentials are valid

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
}

func verifyPlantCreds(w http.ResponseWriter, r *http.Request) bool {
	// Get plant credentials from cookies
	cookie, err := r.Cookie("plant_id")
	if err != nil {
		http.Error(w, "plant_id cookie required", http.StatusUnauthorized)
		return false
	}
	plantID := cookie.Value

	secretCookie, err := r.Cookie("plant_secret")
	if err != nil {
		http.Error(w, "plant_secret cookie required", http.StatusUnauthorized)
		return false
	}
	plantSecret := secretCookie.Value

	// Get the stored hash from the database
	var storedHash string
	err = db.QueryRow("SELECT plant_secret_hash FROM plants WHERE plant_id = ?", plantID).Scan(&storedHash)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid plant credentials", http.StatusUnauthorized)
		return false
	} else if err != nil {
		log.Printf("Error querying plant secret hash: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return false
	}

	// Verify the secret against the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plantSecret))
	if err != nil {
		http.Error(w, "invalid plant credentials", http.StatusUnauthorized)
		return false
	}

	return true
}
