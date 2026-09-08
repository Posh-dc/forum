package backEnd

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {

	conn, err := pgx.Connect(context.Background(), "postgresql://postgres.riwtdppzfdfhycmpjndy:HvJCW0HwQzxrPDkG@aws-1-eu-west-1.pooler.supabase.com:6543/postgres?pgbouncer=true")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(context.Background())
	log.Println("Connected")
}
