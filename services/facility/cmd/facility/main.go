package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"kursovaya_aksp/services/facility/internal/app"
	"kursovaya_aksp/services/facility/internal/grpcapi"
	"kursovaya_aksp/services/facility/internal/store"
	"kursovaya_aksp/services/facility/pkg/grpc/facilitypb"
)

const (
	defaultAddr        = ":8082"
	defaultGRPCAddr    = ":9082"
	defaultDatabaseDSN = "postgres://postgres:postgres@localhost:5432/sporthub?sslmode=disable"
)

func main() {
	databaseURL := getenv("FACILITY_DATABASE_URL", defaultDatabaseDSN)
	pool := connectPool(databaseURL)
	defer pool.Close()

	dataStore := store.New(pool)
	dataStore.SeedDemoData()

	server := app.New(dataStore)
	grpcAddr := getenv("FACILITY_GRPC_ADDR", defaultGRPCAddr)
	go func() {
		if err := serveGRPC(dataStore, grpcAddr); err != nil {
			log.Fatalf("facility gRPC server failed: %v", err)
		}
	}()
	addr := getenv("FACILITY_ADDR", defaultAddr)
	log.Printf("facility service listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Router()); err != nil {
		log.Fatalf("facility service failed: %v", err)
	}
}

func serveGRPC(store *store.Store, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	facilitypb.RegisterFacilityServiceServer(grpcServer, grpcapi.New(store))
	log.Printf("facility gRPC listening on %s", addr)
	return grpcServer.Serve(lis)
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
