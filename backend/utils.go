// `utils.go` has some misc. utility endpoints
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func generateAlphanumeric(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func handleGeneratePlants(w http.ResponseWriter, r *http.Request) {
	// Generate 10 ids in the form of plant_xxxxx where xxxxx is a random 5 digit number
	log.Println("Generating 10 plant IDs and secrets")
	var plantIDs []string
	var plantSecrets []string
	for len(plantIDs) < 10 {
		plantID := fmt.Sprintf("plant_%05d", rand.Int()%100000)
		// check if plantID is already in the database
		var exists bool
		err := db.QueryRow("SELECT plant_id FROM plants WHERE plant_id = ?", plantID).Scan(&exists)
		if err == sql.ErrNoRows {
			plant_secret := generateAlphanumeric(16)
			plant_secret_hash, err := bcrypt.GenerateFromPassword([]byte(plant_secret), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Error hashing plant secret: %v", err)
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}
			// insert plantID and plant_secret_hash into plants table
			_, err = db.Exec("INSERT INTO plants (plant_id, plant_secret_hash) VALUES (?, ?)", plantID, plant_secret_hash)
			if err != nil {
				log.Printf("Error inserting plant: %v", err)
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}
			plantIDs = append(plantIDs, plantID)
			plantSecrets = append(plantSecrets, plant_secret)
		} else if err != nil {
			log.Printf("Error checking plant ID existence: %v", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"plantIds": plantIDs, "plantSecrets": plantSecrets})
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
