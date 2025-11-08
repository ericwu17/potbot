package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/securecookie"
	"github.com/joho/godotenv"
)

var db *sql.DB
var s *securecookie.SecureCookie

const cookieName = "potbot_session"

type User struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func main() {
	// Load .env file if present so DATABASE_USER and DATABASE_PASSWORD
	// (and other env vars) are available via os.Getenv. Use godotenv
	// which is a small, well-tested library for this purpose.
	if err := godotenv.Load(); err != nil {
		log.Printf("godotenv: could not load .env: %v", err)
	}

	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DATABASE_USER")
	cfg.Passwd = os.Getenv("DATABASE_PASSWORD")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "potbot"

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)

	hashKey := os.Getenv("POTBOT_HASH_KEY")
	blockKey := os.Getenv("POTBOT_BLOCK_KEY")
	s = securecookie.New([]byte(hashKey), []byte(blockKey))

	http.HandleFunc("/api/register", withCORS(handleRegister))
	http.HandleFunc("/api/login", withCORS(handleLogin))
	http.HandleFunc("/api/logout", withCORS(handleLogout))
	http.HandleFunc("/api/me", withCORS(handleMe))
	http.HandleFunc("/api/add_plant", withCORS(handleAddPlant))
	http.HandleFunc("/api/get_all_my_plants", withCORS(handleGetAllMyPlants))
	http.HandleFunc("/api/generate_plants", withCORS(handleGeneratePlants))

	// Serve frontend static if built into ./frontend/build
	fs := http.FileServer(http.Dir("../frontend/build"))
	http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	http.ListenAndServe(":"+port, nil)
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Basic CORS for dev. Allow credentials.
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}
	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// insert
	res, err := db.Exec("INSERT INTO users (email, password_hash, username) VALUES (?, ?, ?)", req.Email, string(hash), nullableString(req.Username))
	if err != nil {
		http.Error(w, fmt.Sprintf("db error: %v", err), http.StatusBadRequest)
		return
	}
	id, _ := res.LastInsertId()
	setSessionCookie(w, int(id))
	user := User{UserID: int(id), Email: req.Email, Username: req.Username}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}
	var id int
	var hash string
	var email string
	err := db.QueryRow("SELECT user_id, password_hash, email FROM users WHERE username = ?", req.Username).Scan(&id, &hash, &email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	setSessionCookie(w, id)
	user := User{UserID: id, Email: email, Username: req.Username}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, ok := getSessionUserID(r)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	var u User
	var username sql.NullString
	err := db.QueryRow("SELECT user_id, email, username FROM users WHERE user_id = ?", id).Scan(&u.UserID, &u.Email, &username)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	if username.Valid {
		u.Username = username.String
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// (no local .env parser; using github.com/joho/godotenv instead)

func setSessionCookie(w http.ResponseWriter, userID int) {
	value := map[string]string{"user_id": strconv.Itoa(userID)}
	if encoded, err := s.Encode(cookieName, value); err == nil {
		cookie := &http.Cookie{
			Name:     cookieName,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // set to true in production with HTTPS
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400,
		}
		http.SetCookie(w, cookie)
	} else {
		log.Printf("Error encoding cookie: %v", err)
	}
}

func clearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

func getSessionUserID(r *http.Request) (int, bool) {
	c, err := r.Cookie(cookieName)

	if err != nil {
		return 0, false
	}
	var value map[string]string
	if err = s.Decode(cookieName, c.Value, &value); err != nil {
		return 0, false
	}
	id_str, ok := value["user_id"]
	if !ok {
		return 0, false
	}
	id, err := strconv.Atoi(id_str)
	if err != nil {
		return 0, false
	}
	return id, true
}

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
