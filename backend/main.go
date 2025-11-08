package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/securecookie"
	"github.com/joho/godotenv"
)

var db *sql.DB
var secCookie *securecookie.SecureCookie

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
	secCookie = securecookie.New([]byte(hashKey), []byte(blockKey))

	http.HandleFunc("/api/register", withCORS(handleRegister))
	http.HandleFunc("/api/login", withCORS(handleLogin))
	http.HandleFunc("/api/logout", withCORS(handleLogout))
	http.HandleFunc("/api/me", withCORS(handleMe))
	http.HandleFunc("/api/add_plant", withCORS(handleAddPlant))
	http.HandleFunc("/api/get_all_my_plants", withCORS(handleGetAllMyPlants))
	http.HandleFunc("/api/generate_plants", withCORS(handleGeneratePlants))
	http.HandleFunc("/api/ping", withCORS(handlePing))
	http.HandleFunc("/api/verify_plant_creds", withCORS(handleVerifyPlantCreds))
	http.HandleFunc("/api/plant_log", withCORS(handlePlantLog))

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
