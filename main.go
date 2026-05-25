package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	initTables()

	http.HandleFunc("/signup", LoggingMiddleware(Signup))
	http.HandleFunc("/login", LoggingMiddleware(Login))
	http.HandleFunc("/logout", LoggingMiddleware(AuthMiddleware(Logout)))

	http.HandleFunc("/admin/waiters", LoggingMiddleware(AuthMiddleware(GetWaiters)))
	http.HandleFunc("/admin/stats", LoggingMiddleware(AuthMiddleware(GetWaiterStats)))
	http.HandleFunc("/admin/assign-table", LoggingMiddleware(AuthMiddleware(AssignTable)))
	http.HandleFunc("/admin/delete-waiter", LoggingMiddleware(AuthMiddleware(DeleteWaiter)))

	http.HandleFunc("/waiter/status", LoggingMiddleware(AuthMiddleware(UpdateAvailability)))

	http.HandleFunc("/table/status", LoggingMiddleware(AuthMiddleware(UpdateTableStatus)))
	http.HandleFunc("/tables", LoggingMiddleware(GetTables))
	http.HandleFunc("/queue", LoggingMiddleware(GetQueue))

	http.HandleFunc("/ws", HandleWebSocket)

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/health", HealthCheck)

	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"*",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"*",
		},
		AllowCredentials: true,
	})

	handler := c.Handler(http.DefaultServeMux)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port :", port)

	log.Fatal(http.ListenAndServe(":"+port, handler))
}