package main

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Println(r.Method, r.URL.Path)
		next(w, r)
		log.Println("Done:", time.Since(start))
	}
}
