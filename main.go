package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// Stateless session id based REST API to handle user notes

var db *sql.DB
var sessions = make(map[string]string)

// User represents a user with name, email, and password.
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Note represents a user's note.
type Note struct {
	ID   int    `json:"id"`
	Note string `json:"note"`
}

func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create user and note tables if they don't exist
	createTables()

	r := mux.NewRouter()

	// Endpoint for creating a new user
	r.HandleFunc("/signup", CreateUser).Methods("POST")

	// Endpoint for user login
	r.HandleFunc("/login", Login).Methods("POST")

	// Endpoint for listing all notes created by a user
	r.HandleFunc("/notes", ListNotes).Methods("GET")

	// Endpoint for creating a new note
	r.HandleFunc("/notes", CreateNote).Methods("POST")

	// Endpoint for deleting a note
	r.HandleFunc("/notes", DeleteNote).Methods("DELETE")

	// Start the server
	port := 8795
	fmt.Printf("Server is running on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}

func createTables() {
	// Create the users table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			email TEXT,
			password TEXT
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create the notes table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			note TEXT
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
}

// CreateUser handles user registration.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Insert the user into the database
	_, err := db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", user.Name, user.Email, user.Password)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Login handles user login and returns a session ID.
func Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Check if the user exists and the password matches
	var user User
	err := db.QueryRow("SELECT id, name FROM users WHERE email = ? AND password = ?", loginRequest.Email, loginRequest.Password).Scan(&user.ID, &user.Name)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := generateSessionID()
	sessions[sessionID] = loginRequest.Email

	response := struct {
		SID string `json:"sid"`
	}{SID: sessionID}
	json.NewEncoder(w).Encode(response)
}

// ListNotes returns all notes created by a user.
func ListNotes(w http.ResponseWriter, r *http.Request) {
	var noteRequest struct {
		SID string `json:"sid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&noteRequest); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	email, err := isValidSessionID(noteRequest.SID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the user's ID based on the session ID
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	var userNotes []Note
	rows, err := db.Query("SELECT id, note FROM notes WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Error listing notes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Note)
		if err != nil {
			http.Error(w, "Error reading notes", http.StatusInternalServerError)
			return
		}
		userNotes = append(userNotes, note)
	}

	response := struct {
		Notes []Note `json:"notes"`
	}{Notes: userNotes}
	json.NewEncoder(w).Encode(response)
}

// CreateNote creates a new note for a user.
func CreateNote(w http.ResponseWriter, r *http.Request) {
	var noteRequest struct {
		SID  string `json:"sid"`
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&noteRequest); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	email, err := isValidSessionID(noteRequest.SID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the user's ID based on the session ID
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Insert the note into the database
	_, err = db.Exec("INSERT INTO notes (user_id, note) VALUES (?, ?)", userID, noteRequest.Note)
	if err != nil {
		http.Error(w, "Error creating note", http.StatusInternalServerError)
		return
	}

	var noteID int
	err = db.QueryRow("SELECT max(id) FROM notes").Scan(&noteID)
	if err != nil {
		http.Error(w, "Note ID not found", http.StatusUnauthorized)
		return
	}

	response := struct {
		ID int `json:"id"`
	}{ID: noteID}
	json.NewEncoder(w).Encode(response)
}

// DeleteNote deletes a note.
func DeleteNote(w http.ResponseWriter, r *http.Request) {
	var deleteRequest struct {
		SID string `json:"sid"`
		ID  int    `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&deleteRequest); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	email, err := isValidSessionID(deleteRequest.SID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the user's ID based on the session ID
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Delete the note if it belongs to the user
	_, err = db.Exec("DELETE FROM notes WHERE id = ? AND user_id = ?", deleteRequest.ID, userID)
	if err != nil {
		http.Error(w, "Error deleting note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// check if session is valid
func isValidSessionID(sid string) (string, error) {
	if email, ok := sessions[sid]; ok {
		fmt.Println(sid, email)
		return email, nil
	} else {
		fmt.Println("Invalid sid", sid)
		return "", fmt.Errorf("Invalid sid")
	}
}
