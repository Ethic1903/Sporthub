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

	"kursovaya_aksp/services/identity/internal/app"
	"kursovaya_aksp/services/identity/internal/grpcapi"
	"kursovaya_aksp/services/identity/internal/store"
	"kursovaya_aksp/services/identity/pkg/grpc/identitypb"
)

const (
	defaultAddr        = ":8081"
	defaultGRPCAddr    = ":9081"
	defaultDatabaseDSN = "postgres://postgres:postgres@localhost:5432/sporthub?sslmode=disable"
)

func main() {
	databaseURL := getenv("IDENTITY_DATABASE_URL", defaultDatabaseDSN)
	pool := connectPool(databaseURL)
	defer pool.Close()

	dataStore := store.New(pool)
	dataStore.SeedDemoUsers()

	server := app.New(dataStore)
	grpcAddr := getenv("IDENTITY_GRPC_ADDR", defaultGRPCAddr)
	go func() {
		if err := serveGRPC(dataStore, grpcAddr); err != nil {
			log.Fatalf("identity gRPC server failed: %v", err)
		}
	}()
	addr := getenv("IDENTITY_ADDR", defaultAddr)
	log.Printf("identity service listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Router()); err != nil {
		log.Fatalf("identity service failed: %v", err)
	}
}

func serveGRPC(store *store.Store, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	identitypb.RegisterIdentityServiceServer(grpcServer, grpcapi.New(store))
	log.Printf("identity gRPC listening on %s", addr)
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
