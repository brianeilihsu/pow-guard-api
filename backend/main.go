package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

// Challenge storage with expiration
type Challenge struct {
	Challenge  string
	Difficulty int
	ExpiresAt  time.Time
}

var (
	challenges = make(map[string]Challenge)
	challengeMu sync.RWMutex

	// PoW configuration
	powDifficulty = 4 // Number of leading zeros required in hash
	challengeTTL = 5 * time.Minute
)

// User represents a user in the system
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Challenge string `json:"challenge"`
	Nonce     string `json:"nonce"`
}

// ChallengeResponse represents the challenge response
type ChallengeResponse struct {
	Challenge  string `json:"challenge"`
	Difficulty int    `json:"difficulty"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	// Initialize database connection
	var err error
	connStr := "host=localhost port=5432 user=powuser password=powpass dbname=powdb sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Successfully connected to database")

	// Start challenge cleanup routine
	go cleanupExpiredChallenges()

	router := mux.NewRouter()

	// Add CORS middleware
	router.Use(corsMiddleware)

	router.HandleFunc("/api/challenge", getChallengeHandler).Methods("GET")
	router.HandleFunc("/api/register", registerHandler).Methods("POST")
	router.HandleFunc("/health", healthHandler).Methods("GET")

	log.Println("Server starting on :8080")
	log.Println("PoW difficulty:", powDifficulty, "leading zeros")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// cleanupExpiredChallenges removes expired challenges periodically
func cleanupExpiredChallenges() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		challengeMu.Lock()
		now := time.Now()
		for id, challenge := range challenges {
			if now.After(challenge.ExpiresAt) {
				delete(challenges, id)
			}
		}
		challengeMu.Unlock()
	}
}

// generateChallenge creates a new PoW challenge
func generateChallenge() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// getChallengeHandler returns a new PoW challenge
func getChallengeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	challenge, err := generateChallenge()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to generate challenge",
		})
		return
	}

	// Store challenge
	challengeMu.Lock()
	challenges[challenge] = Challenge{
		Challenge:  challenge,
		Difficulty: powDifficulty,
		ExpiresAt:  time.Now().Add(challengeTTL),
	}
	challengeMu.Unlock()

	log.Printf("Generated challenge: %s (expires in %v)", challenge[:16]+"...", challengeTTL)

	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Challenge generated",
		Data: ChallengeResponse{
			Challenge:  challenge,
			Difficulty: powDifficulty,
		},
	})
}

// verifyPoW verifies the proof of work
func verifyPoW(challenge, nonce string, difficulty int) bool {
	// Compute hash of challenge + nonce
	data := challenge + nonce
	hash := sha256.Sum256([]byte(data))
	hashHex := hex.EncodeToString(hash[:])

	// Check if hash has required number of leading zeros
	requiredPrefix := ""
	for i := 0; i < difficulty; i++ {
		requiredPrefix += "0"
	}

	return len(hashHex) >= difficulty && hashHex[:difficulty] == requiredPrefix
}

// corsMiddleware handles CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Server is running",
	})
}

// registerHandler handles user registration with PoW protection
func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Verify PoW
	challengeMu.RLock()
	storedChallenge, exists := challenges[req.Challenge]
	challengeMu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid or expired challenge",
		})
		return
	}

	if time.Now().After(storedChallenge.ExpiresAt) {
		challengeMu.Lock()
		delete(challenges, req.Challenge)
		challengeMu.Unlock()

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Challenge expired",
		})
		return
	}

	if !verifyPoW(req.Challenge, req.Nonce, storedChallenge.Difficulty) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid proof of work",
		})
		return
	}

	// Remove used challenge
	challengeMu.Lock()
	delete(challenges, req.Challenge)
	challengeMu.Unlock()

	log.Printf("PoW verified for challenge: %s...", req.Challenge[:16])

	// Validate input
	if err := validateRegisterRequest(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Check if user already exists
	var existingID int
	err := db.QueryRow("SELECT id FROM users WHERE username = $1", req.Username).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(Response{
				Success: false,
				Message: "Username already exists",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Database error",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to process password",
		})
		return
	}

	// Insert user into database
	var user User
	err = db.QueryRow(
		"INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id, username, email, created_at",
		req.Username,
		string(hashedPassword),
		req.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err != nil {
		log.Println("Database insert error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to create user",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}

// validateRegisterRequest validates the registration request
func validateRegisterRequest(req RegisterRequest) error {
	if req.Username == "" {
		return &ValidationError{"Username is required"}
	}

	if len(req.Username) < 3 {
		return &ValidationError{"Username must be at least 3 characters"}
	}

	if req.Password == "" {
		return &ValidationError{"Password is required"}
	}

	if len(req.Password) < 6 {
		return &ValidationError{"Password must be at least 6 characters"}
	}

	if req.Email == "" {
		return &ValidationError{"Email is required"}
	}

	// Simple email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return &ValidationError{"Invalid email format"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
