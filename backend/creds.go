// `creds.go` contains the code for handling user authentication
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

const cookieName = "potbot_session"

func setSessionCookie(w http.ResponseWriter, userID int) {
	value := map[string]string{"user_id": strconv.Itoa(userID)}
	if encoded, err := secCookie.Encode(cookieName, value); err == nil {
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
	if err = secCookie.Decode(cookieName, c.Value, &value); err != nil {
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
