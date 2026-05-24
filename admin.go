package main

import (
	"encoding/json"
	"net/http"
)

func GetWaiters(w http.ResponseWriter, r *http.Request) {

	var list []User

	mu.Lock()
	for _, u := range users {
		if u.Role == "waiter" {
			list = append(list, u)
		}
	}
	mu.Unlock()

	json.NewEncoder(w).Encode(list)
}

func GetWaiterStats(w http.ResponseWriter, r *http.Request) {

	total := 0
	available := 0

	mu.Lock()
	for _, u := range users {
		if u.Role == "waiter" {
			total++
			if u.Available {
				available++
			}
		}
	}
	mu.Unlock()

	json.NewEncoder(w).Encode(map[string]int{
		"total":     total,
		"available": available,
	})

}

func DeleteWaiter(w http.ResponseWriter, r *http.Request) {

	//  ONLY ADMIN ALLOWED
	if r.Header.Get("user_id") != "ADMIN" {
		http.Error(w, "Only admin allowed", 401)
		return
	}

	waiterID := r.URL.Query().Get("emp_id")
	if waiterID == "" {
		http.Error(w, "emp_id required", 400)
		return
	}

	// DELETE USER
	mu.Lock()
	_, exists := users[waiterID]
	if !exists {
		mu.Unlock()
		http.Error(w, "Waiter not found", 404)
		return
	}
	delete(users, waiterID)
	mu.Unlock()

	// REMOVE FROM TABLES
	tableMu.Lock()
	for id, t := range tables {
		if t.WaiterID == waiterID {
			t.WaiterID = getAvailableWaiter()
			tables[id] = t
		}
	}
	tableMu.Unlock()

	// WEBSOCKET
	broadcast("Waiter deleted")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Waiter deleted successfully",
	})
}
