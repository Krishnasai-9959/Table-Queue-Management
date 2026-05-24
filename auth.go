package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var mu sync.Mutex
var users = make(map[string]User)

const adminEmail = "admin@restaurant.com"
const adminPassword = "Admin123"

type User struct {
	EmpID     string `json:"emp_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	Available bool   `json:"available"`
}

// EMAIL VALIDATION
func isValidEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(regex).MatchString(email)
}

// PASSWORD VALIDATION
func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	var upper, lower, digit bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			upper = true
		case c >= 'a' && c <= 'z':
			lower = true
		case c >= '0' && c <= '9':
			digit = true
		}
	}
	return upper && lower && digit
}

// SIGNUP
func Signup(w http.ResponseWriter, r *http.Request) {

	var input User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if !isValidEmail(input.Email) {
		http.Error(w, "Invalid email", 400)
		return
	}

	if !isStrongPassword(input.Password) {
		http.Error(w, "Weak password", 400)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	id := uuid.New().String()

	mu.Lock()
	users[id] = User{id, input.Name, input.Email, string(hash), "waiter", true}
	mu.Unlock()

	token, _ := GenerateJWT(id, "waiter")

	json.NewEncoder(w).Encode(map[string]string{
		"user_id": id,
		"role":    "waiter",
		"token":   token,
	})
}

// LOGIN
func Login(w http.ResponseWriter, r *http.Request) {

	var input User
	json.NewDecoder(r.Body).Decode(&input)

	// ADMIN
	if input.Email == adminEmail && input.Password == adminPassword {
		token, _ := GenerateJWT("ADMIN", "admin")
		json.NewEncoder(w).Encode(map[string]string{
			"user_id": "ADMIN",
			"role":    "admin",
			"token":   token,
		})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for id, u := range users {
		if u.Email == input.Email &&
			bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(input.Password)) == nil {

			token, _ := GenerateJWT(id, "waiter")

			json.NewEncoder(w).Encode(map[string]string{
				"user_id": id,
				"role":    "waiter",
				"token":   token,
			})
			return
		}
	}

	http.Error(w, "Invalid credentials", 401)
}

// LOGOUT
func Logout(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "logout"})
}

// HELPER
func getAvailableWaiter() string {
	mu.Lock()
	defer mu.Unlock()
	for _, u := range users {
		if u.Available {
			return u.EmpID
		}
	}
	return ""
}
