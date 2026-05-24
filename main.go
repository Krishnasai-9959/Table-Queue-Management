package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	initTables()

	// AUTH
	http.HandleFunc("/signup", LoggingMiddleware(Signup))
	http.HandleFunc("/login", LoggingMiddleware(Login))
	http.HandleFunc("/logout", LoggingMiddleware(AuthMiddleware(Logout)))

	// ADMIN
	http.HandleFunc("/admin/waiters", LoggingMiddleware(AuthMiddleware(GetWaiters)))
	http.HandleFunc("/admin/stats", LoggingMiddleware(AuthMiddleware(GetWaiterStats)))
	http.HandleFunc("/admin/assign-table", LoggingMiddleware(AuthMiddleware(AssignTable)))
	http.HandleFunc("/admin/delete-waiter", LoggingMiddleware(AuthMiddleware(DeleteWaiter)))

	// WAITER
	http.HandleFunc("/waiter/status", LoggingMiddleware(AuthMiddleware(UpdateAvailability)))

	// TABLE
	http.HandleFunc("/table/status", LoggingMiddleware(AuthMiddleware(UpdateTableStatus)))
	http.HandleFunc("/tables", LoggingMiddleware(GetTables))
	http.HandleFunc("/queue", LoggingMiddleware(GetQueue))

	// WEBSOCKET
	http.HandleFunc("/ws", HandleWebSocket)

	// METRICS
	http.Handle("/metrics", promhttp.Handler())

	// 🔥 SERVE FRONTEND
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	handler := c.Handler(http.DefaultServeMux)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
