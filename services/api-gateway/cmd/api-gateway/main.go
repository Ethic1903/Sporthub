package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"kursovaya_aksp/services/api-gateway/internal/app"
)

const (
	defaultGatewayAddr      = ":8080"
	identityBaseDefault     = "http://localhost:8081"
	facilityBaseDefault     = "http://localhost:8082"
	bookingBaseDefault      = "http://localhost:8083"
	notificationBaseDefault = "http://localhost:8084"
)

func main() {
	cfg := app.Config{
		IdentityURL:     getenv("IDENTITY_URL", identityBaseDefault),
		FacilityURL:     getenv("FACILITY_URL", facilityBaseDefault),
		BookingURL:      getenv("BOOKING_URL", bookingBaseDefault),
		NotificationURL: getenv("NOTIFICATION_URL", notificationBaseDefault),
		Client:          &http.Client{Timeout: 10 * time.Second},
	}

	server := app.New(cfg)
	addr := getenv("GATEWAY_ADDR", defaultGatewayAddr)
	log.Printf("api gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Router()); err != nil {
		log.Fatalf("gateway failed: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
