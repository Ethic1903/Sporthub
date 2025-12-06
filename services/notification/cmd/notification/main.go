package main

import (
	"log"
	"net/http"
	"os"

	"kursovaya_aksp/services/notification/internal/app"
	"kursovaya_aksp/services/notification/internal/store"
)

const defaultAddr = ":8084"

func main() {
	dataStore := store.New()
	server := app.New(dataStore)

	addr := getenv("NOTIFICATION_ADDR", defaultAddr)
	log.Printf("notification service listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Router()); err != nil {
		log.Fatalf("notification service failed: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
