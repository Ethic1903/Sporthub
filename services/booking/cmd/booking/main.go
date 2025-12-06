package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"kursovaya_aksp/services/booking/internal/app"
	"kursovaya_aksp/services/booking/internal/store"
	"kursovaya_aksp/services/facility/pkg/grpc/facilitypb"
	"kursovaya_aksp/services/identity/pkg/grpc/identitypb"
)

const (
	defaultAddr             = ":8083"
	defaultFacilityGRPCAddr = "localhost:9082"
	defaultIdentityGRPCAddr = "localhost:9081"
	defaultDatabaseDSN      = "postgres://postgres:postgres@localhost:5432/sporthub?sslmode=disable"
)

func main() {
	databaseURL := getenv("BOOKING_DATABASE_URL", defaultDatabaseDSN)
	pool := connectPool(databaseURL)
	defer pool.Close()

	dataStore := store.New(pool)
	facilityConn := dialGRPC("FACILITY_GRPC_ADDR", defaultFacilityGRPCAddr)
	defer facilityConn.Close()
	identityConn := dialGRPC("IDENTITY_GRPC_ADDR", defaultIdentityGRPCAddr)
	defer identityConn.Close()
	facilityClient := facilitypb.NewFacilityServiceClient(facilityConn)
	identityClient := identitypb.NewIdentityServiceClient(identityConn)
	server := app.New(dataStore, facilityClient, identityClient)

	addr := getenv("BOOKING_ADDR", defaultAddr)
	log.Printf("booking service listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Router()); err != nil {
		log.Fatalf("booking service failed: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func connectPool(url string) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatalf("failed to init database pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	return pool
}

func dialGRPC(envKey, fallback string) *grpc.ClientConn {
	addr := getenv(envKey, fallback)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial %s (%s): %v", envKey, addr, err)
	}
	return conn
}
