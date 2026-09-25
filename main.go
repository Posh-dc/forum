package main

import (
	"context"
	"fmt"
	"log"
	"main/backEnd"
	"time"

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
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	ctxx, cancle := context.WithTimeout(context.Background(), time.Hour)
	defer cancle()
	godotenv.Load()

	pool, dbError := databaseConnect(ctx)

	defer pool.Close()

	if dbError != nil {
		log.Fatal(dbError)
	}

	dbConnect := &backEnd.DBstruct{
		DB: pool,
	}

	// backEnd.SendMail()

	// emails, err := backEnd.GetPendingEmails(ctx, pool)

	// if err != nil {
	// 	fmt.Printf("Getpending email error : %w", err)
	// 	log.Fatal()
	// }

	// fmt.Printf("%+v", emails)

	backEnd.ProcessPendingEmails(ctxx, pool)

	mux := http.NewServeMux()

	// serving front end starts here
	fs := http.FileServer(http.Dir("frontend"))
	mux.Handle("/frontend/", http.StripPrefix("/frontend/", fs))
	// serving front end starts here

	// home page router
	mux.HandleFunc("/", backEnd.HomeHandler)

	// registration and loging routers
	mux.HandleFunc("/onboarding", backEnd.OnboardingHandler)
	mux.HandleFunc("/userName", dbConnect.UsernameAvailabilityHandler)
	mux.HandleFunc("/register", dbConnect.CreateAccountHandler)

	server := &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: mux,
	}

	fmt.Printf("server running on http://localhost:%v\n", server.Addr)
	server.ListenAndServe()

}
