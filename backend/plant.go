// `plant.go` contains all the endpoints that a plant interacts with
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var validLogTypes = []string{"light", "temp", "moisture"}

func handleVerifyPlantCreds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if ok, _ := verifyPlantCreds(w, r); !ok {
		return
	}
	// If we get here, credentials are valid

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
}

func verifyPlantCreds(w http.ResponseWriter, r *http.Request) (bool, string) {
	// Get plant credentials from cookies
	cookie, err := r.Cookie("plant_id")
	if err != nil {
		http.Error(w, "plant_id cookie required", http.StatusUnauthorized)
		return false, ""
	}
	plantID := cookie.Value

	secretCookie, err := r.Cookie("plant_secret")
	if err != nil {
		http.Error(w, "plant_secret cookie required", http.StatusUnauthorized)
		return false, ""
	}
	plantSecret := secretCookie.Value

	// Get the stored hash from the database
	var storedHash string
	err = db.QueryRow("SELECT plant_secret_hash FROM plants WHERE plant_id = ?", plantID).Scan(&storedHash)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid plant credentials", http.StatusUnauthorized)
		return false, ""
	} else if err != nil {
		log.Printf("Error querying plant secret hash: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return false, ""
	}

	// Verify the secret against the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plantSecret))
	if err != nil {
		http.Error(w, "invalid plant credentials", http.StatusUnauthorized)
		return false, ""
	}

	return true, plantID
}

// Request body for logging a plant value
type plantLogRequest struct {
	LogType  string  `json:"logType"`
	LogValue float64 `json:"logValue"`
}

// handlePlantLog allows a plant (authenticated via cookies) to log a value of a specific type.
// Expects JSON body: { "logType": "<type>", "logValue": <number> }
// Uses current server time for log_time.
// logType must be one of validLogTypes
func handlePlantLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ok, plantID := verifyPlantCreds(w, r)
	if !ok {
		return
	}

	var req plantLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !slices.Contains(validLogTypes, req.LogType) {
		http.Error(w, "invalid log type", http.StatusBadRequest)
		return
	}

	// Insert into plant_logs using current server time
	_, err := db.Exec(
		"INSERT INTO plant_logs (plant_id, log_type, log_time, log_value) VALUES (?, ?, ?, ?)",
		plantID, req.LogType, time.Now(), req.LogValue,
	)
	if err != nil {
		log.Printf("error inserting plant log: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
