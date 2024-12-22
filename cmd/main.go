package main

import (
	"exiftool-wrapper/server"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/tags", server.HandleTags)

	println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		println("Server error:", err.Error())
	}
}
