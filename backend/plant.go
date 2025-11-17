// `plant.go` contains all the endpoints that a plant interacts with
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

// handleFetchCommands allows an authenticated plant (via cookies) to fetch
// all pending commands queued for it. It returns a JSON array of strings
// and clears the pending commands for that plant.
func handleFetchCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ok, plantID := verifyPlantCreds(w, r)
	if !ok {
		return
	}

	var cmds []string = make([]string, 0)
	if pendingCommands != nil {
		if c, exists := pendingCommands[plantID]; exists {
			cmds = c
		}
		// clear commands for this plant
		delete(pendingCommands, plantID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cmds)
}

// Request body for plant notifications
type plantNotifyRequest struct {
	NotificationType string `json:"notificationType"`
}

// handlePlantNotify is called by a plant (authenticated via cookies) to notify
// the owner about an event. It expects JSON body: { "notificationType": "xxxxx" }
func handlePlantNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ok, plantID := verifyPlantCreds(w, r)
	if !ok {
		return
	}

	var req plantNotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.NotificationType == "" {
		http.Error(w, "notification_type is required", http.StatusBadRequest)
		return
	}

	// Lookup owner's email for the plant
	var ownerEmail sql.NullString
	err := db.QueryRow("SELECT u.email FROM users u JOIN plants p ON p.user_id = u.user_id WHERE p.plant_id = ?", plantID).Scan(&ownerEmail)
	if err == sql.ErrNoRows || !ownerEmail.Valid {
		http.Error(w, "plant has no associated user", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Printf("Error querying owner email: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// Build subject and body based on notification type
	subject := fmt.Sprintf("Potbot notification for %s", plantID)
	body := fmt.Sprintf("Your plant (ID: %s) sent notification: %s", plantID, req.NotificationType)
	switch req.NotificationType {
	case "FALLEN":
		subject = "Your plant has fallen over"
		body = fmt.Sprintf("Hi — your plant (ID: %s) appears to have fallen over. Please check on it.", plantID)
	}

	if err := sendEmail(ownerEmail.String, subject, body); err != nil {
		log.Printf("error sending notification email: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "notified"})
}
