package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	addr := ":3000"
	log.Printf("Server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}