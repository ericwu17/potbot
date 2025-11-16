// `user.go` contains all the endpoints that a user interacts with

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"time"
)

func handleGetAllMyPlants(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getSessionUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Query plants for the user. Return plantName and type to match frontend usage.
	rows, err := db.Query("SELECT plant_name, plant_id, plant_type FROM plants WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("Error querying plants: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Plant struct {
		PlantName string `json:"plantName"`
		PlantID   string `json:"plantID"`
		Type      string `json:"type"`
	}

	var plants []Plant
	for rows.Next() {
		var pname sql.NullString
		var ptype sql.NullString
		var pid sql.NullString
		if err := rows.Scan(&pname, &pid, &ptype); err != nil {
			log.Printf("Error scanning plant row: %v", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		p := Plant{}
		if pname.Valid {
			p.PlantName = pname.String
		}
		if ptype.Valid {
			p.Type = ptype.String
		}
		if pid.Valid {
			p.PlantID = pid.String
		}
		plants = append(plants, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plants)
}

func handleAddPlant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	userID, ok := getSessionUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		PlantID   string `json:"plantId"`
		PlantName string `json:"plantName"`
		Type      string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.PlantID == "" || req.Type == "" {
		http.Error(w, "plantId and type are required", http.StatusBadRequest)
		return
	}

	// Check if the plant ID exists and is not already associated with a user
	var existingUserID sql.NullInt64
	err := db.QueryRow("SELECT user_id FROM plants WHERE plant_id = ?", req.PlantID).Scan(&existingUserID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid plant ID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Printf("Error checking plant ID: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if existingUserID.Valid {
		http.Error(w, "plant ID already associated with a user", http.StatusBadRequest)
		return
	}

	// Associate plant with user
	_, err = db.Exec("UPDATE plants SET user_id = ?, plant_type = ?, plant_name = ? WHERE plant_id = ?", userID, req.Type, req.PlantName, req.PlantID)
	if err != nil {
		log.Printf("Error associating plant with user: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Request body for issuing a command to a plant
type issueCommandRequest struct {
	PlantID string `json:"plantId"`
	Command string `json:"command"`
}

// handleIssueCommand allows an authenticated user to enqueue a command for one of their plants.
// Expects POST JSON body: { "plantId": "<id>", "command": "<string>" }
func handleIssueCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getSessionUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req issueCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.PlantID == "" {
		http.Error(w, "plantId is required", http.StatusBadRequest)
		return
	}

	// Verify ownership
	var plantOwnerID int
	err := db.QueryRow("SELECT user_id FROM plants WHERE plant_id = ?", req.PlantID).Scan(&plantOwnerID)
	if err == sql.ErrNoRows {
		http.Error(w, "plant not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error checking plant ownership: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if plantOwnerID != userID {
		http.Error(w, "you do not own this plant", http.StatusForbidden)
		return
	}

	// Enqueue command
	if pendingCommands == nil {
		pendingCommands = make(map[string][]string)
	}
	// Check that the plantID key exists in the pendingCommands map
	if _, exists := pendingCommands[req.PlantID]; !exists {
		pendingCommands[req.PlantID] = []string{}
	}

	pendingCommands[req.PlantID] = append(pendingCommands[req.PlantID], req.Command)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
}

type PlantLogsRequest struct {
	PlantID   string    `json:"plantID"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type PlantLogEntry struct {
	Val  float64   `json:"val"`
	Time time.Time `json:"time"`
}

// handleGetPlantLogs retrieves sensor logs for a specific plant within a date range.
// It verifies that the requesting user owns the plant before returning any data.
//
// Expects a POST request with JSON body:
//
//	{
//	    "plantID": "string",      // ID of the plant to get logs for
//	    "startDate": "time",      // Start of date range (RFC3339 format)
//	    "endDate": "time"         // End of date range (RFC3339 format)
//	}
func handleGetPlantLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	userID, ok := getSessionUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req PlantLogsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.PlantID == "" {
		http.Error(w, "plantId is required", http.StatusBadRequest)
		return
	}

	// Check if the user owns the plant
	var plantOwnerID int
	err := db.QueryRow("SELECT user_id FROM plants WHERE plant_id = ?", req.PlantID).Scan(&plantOwnerID)
	if err == sql.ErrNoRows {
		http.Error(w, "plant not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error checking plant ownership: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if plantOwnerID != userID {
		http.Error(w, "you do not own this plant", http.StatusForbidden)
		return
	}

	// Query plant logs for the specified date range
	rows, err := db.Query(
		"SELECT log_type, log_value, log_time FROM plant_logs WHERE plant_id = ? AND log_time BETWEEN ? AND ? ORDER BY log_time DESC",
		req.PlantID, req.StartDate, req.EndDate,
	)
	if err != nil {
		log.Printf("Error querying plant logs: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := map[string][]PlantLogEntry{}
	for _, t := range validLogTypes {
		result[t] = make([]PlantLogEntry, 0)
	}

	for rows.Next() {
		var logEntry PlantLogEntry

		var logTimeStr string

		var type_ string
		if err := rows.Scan(&type_, &logEntry.Val, &logTimeStr); err != nil {
			log.Printf("Error scanning plant log row: %v", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		logEntry.Time, err = time.Parse("2006-01-02 15:04:05", logTimeStr)
		if err != nil {
			log.Printf("Unable to parse time string: %s", logTimeStr)
			continue
		}

		if !slices.Contains(validLogTypes, type_) {
			log.Printf("Unknown log type in database: %v", type_)
			continue
		}

		result[type_] = append(result[type_], logEntry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
