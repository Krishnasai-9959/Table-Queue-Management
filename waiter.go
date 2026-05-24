package main

import (
	"encoding/json"
	"net/http"
)

func UpdateAvailability(w http.ResponseWriter, r *http.Request) {

	id := r.Header.Get("user_id")

	var input struct {
		Available bool `json:"available"`
	}

	json.NewDecoder(r.Body).Decode(&input)

	mu.Lock()
	user := users[id]
	user.Available = input.Available
	users[id] = user
	mu.Unlock()

	json.NewEncoder(w).Encode(map[string]string{
		"message": "updated",
	})
}
