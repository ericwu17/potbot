// `user.go` contains all the endpoints that a user interacts with

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
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
	rows, err := db.Query("SELECT plant_name, plant_type FROM plants WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("Error querying plants: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Plant struct {
		PlantName string `json:"plantName"`
		Type      string `json:"type"`
	}

	var plants []Plant
	for rows.Next() {
		var pname sql.NullString
		var ptype sql.NullString
		if err := rows.Scan(&pname, &ptype); err != nil {
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
