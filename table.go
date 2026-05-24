package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Table struct {
	ID       int    `json:"id"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
	WaiterID string `json:"waiter_id"`
}

type Queue struct {
	Customers int `json:"customers"`
}

var tableMu sync.Mutex
var tables = make(map[int]Table)
var queue []Queue

//INIT

func initTables() {
	tables[1] = Table{1, 2, "free", ""}
	tables[2] = Table{2, 4, "free", ""}
	tables[3] = Table{3, 4, "free", ""}
	tables[4] = Table{4, 6, "free", ""}
	tables[5] = Table{5, 10, "free", ""}
}

//  ASSIGN TABLE

func AssignTable(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Customers int `json:"customers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	waiterID := getAvailableWaiter()

	tableMu.Lock()
	defer tableMu.Unlock()

	for id, t := range tables {

		//  QUEUE: ASSIGN IF CAPACITY IS SUFFICIENT
		if t.Status == "free" && t.Capacity >= req.Customers {

			t.Status = "occupied"
			t.WaiterID = waiterID
			tables[id] = t

			//  WEBSOCKET
			broadcast("Table assigned")

			json.NewEncoder(w).Encode(map[string]interface{}{
				"table_id":  id,
				"waiter_id": waiterID,
			})
			return
		}
	}

	//  No exact table → add to queue
	queue = append(queue, Queue{req.Customers})

	// WEBSOCKET
	broadcast("Added to queue")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Added to queue",
	})
}

//  UPDATE TABLE STATUS

func UpdateTableStatus(w http.ResponseWriter, r *http.Request) {

	var req struct {
		TableID int    `json:"table_id"`
		Status  string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	waiterID := getAvailableWaiter()

	tableMu.Lock()
	defer tableMu.Unlock()

	t, exists := tables[req.TableID]
	if !exists {
		http.Error(w, "Table not found", 404)
		return
	}

	t.Status = req.Status

	// TIMER FOR CLEANING STATUS
	if req.Status == "cleaning" {
		go func(tableID int) {
			time.Sleep(1 * time.Minute)

			tableMu.Lock()
			defer tableMu.Unlock()

			// Check if it's still cleaning
			tbl, stillExists := tables[tableID]
			if !stillExists || tbl.Status != "cleaning" {
				return
			}

			// Find matching queue person
			matched := false
			wID := getAvailableWaiter()

			if len(queue) > 0 {
				for i, q := range queue {
					if q.Customers <= tbl.Capacity {
						tbl.Status = "occupied"
						tbl.WaiterID = wID

						// remove from queue
						queue = append(queue[:i], queue[i+1:]...)
						matched = true
						break
					}
				}
			}

			if !matched {
				tbl.Status = "free"
			}
			
			tables[tableID] = tbl

			if matched {
				broadcast(fmt.Sprintf("Admin: Table %d cleaned, auto-assigning next group", tableID))
			}
			broadcast("Table updated")

		}(req.TableID)
	}

	//  QUEUE MATCH (FCFS)
	if req.Status == "free" && len(queue) > 0 {

		for i, q := range queue {

			// ASSIGN IF CAPACITY SUFFICIENT
			if q.Customers <= t.Capacity {

				t.Status = "occupied"
				t.WaiterID = waiterID

				// remove from queue
				queue = append(queue[:i], queue[i+1:]...)
				break
			}
		}
	}

	tables[req.TableID] = t

	//  WEBSOCKET
	broadcast("Table updated")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Table updated",
	})
}

//  GET TABLES

func GetTables(w http.ResponseWriter, r *http.Request) {

	tableMu.Lock()
	defer tableMu.Unlock()

	var result []Table
	for _, t := range tables {
		result = append(result, t)
	}

	json.NewEncoder(w).Encode(result)
}

// GET QUEUE

func GetQueue(w http.ResponseWriter, r *http.Request) {

	tableMu.Lock()
	defer tableMu.Unlock()

	json.NewEncoder(w).Encode(queue)
}
