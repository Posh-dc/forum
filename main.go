package main

import (
	"context"
	"fmt"
	"log"
	"main/backEnd"

	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func databaseConnect(ctx context.Context) (*pgxpool.Pool, error) {

	config, configErr := pgxpool.ParseConfig(os.Getenv("SUPABASE_DATABASE_URL"))

	if configErr != nil {
		return nil, fmt.Errorf("conFigErr: \v", configErr)
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	pool, conDbErr := pgxpool.NewWithConfig(ctx, config)

	if conDbErr != nil {
		return nil, fmt.Errorf("conDbErr: \v", conDbErr)
	}
	err := pool.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully Connected to Database")

	return pool, nil
}

func main() {
	ctx := context.Background()

	godotenv.Load()

	pool, dbError := databaseConnect(ctx)

	defer pool.Close()

	if dbError != nil {
		log.Fatal(dbError)
	}

	dbConnect := &backEnd.DBstruct{
		DB: pool,
	}

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	go backEnd.StartEmailWorker(workerCtx, pool)

	mux := http.NewServeMux()

	// serving front end starts here
	fs := http.FileServer(http.Dir("frontend"))
	mux.Handle("/frontend/", http.StripPrefix("/frontend/", fs))
	// serving front end starts here

	// home page router
	mux.HandleFunc("/", backEnd.HomeHandler)

	// registration and loging routers
	mux.HandleFunc("GET /onboarding", backEnd.OnboardingHandler)
	mux.HandleFunc("POST /userName", dbConnect.UsernameAvailabilityHandler)
	mux.HandleFunc("POST /register", dbConnect.CreateAccountHandler)

	// email verification routers
	mux.HandleFunc("GET /verify-email", dbConnect.VerifyEmailPageHandler)
	mux.HandleFunc("POST /verify-email", dbConnect.VerifyEmailHandler)
	server := &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: mux,
	}

	fmt.Printf("server running on http://localhost:%v\n", server.Addr)
	server.ListenAndServe()

}
