package main

import (
	"context"
	"fmt"
	"log"
	"main/backEnd"

	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func databaseConnect() {

	conn, err := pgx.Connect(context.Background(), os.Getenv("SUPABASE_DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(context.Background())
	fmt.Println("Connected")
}

func main() {

	godotenv.Load()

	databaseConnect()

	mux := http.NewServeMux()

	// serving front end starts here
	fs := http.FileServer(http.Dir("frontend"))
	mux.Handle("/frontend/", http.StripPrefix("/frontend/", fs))
	// serving front end starts here

	// home page router
	mux.HandleFunc("/", backEnd.HomeHandler)

	server := &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: mux,
	}

	fmt.Printf("server running on http://localhost:%v\n", server.Addr)
	server.ListenAndServe()

}
