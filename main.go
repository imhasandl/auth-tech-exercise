package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/imhasandl/auth-tech-exercise/cmd/handlers"
	"github.com/imhasandl/auth-tech-exercise/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Can not load .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Set port in .env file")
	}

	tokenSecret := os.Getenv("TOKEN_SECRET")
	if tokenSecret == "" {
		log.Fatal("Set token secret in .env file")
	}

	db_url := os.Getenv("DB_URL")
	if db_url == "" {
		log.Fatal("Set database url in .env file")
	}

	db, err := database.InitDatabase(db_url)
	if err != nil {
		log.Fatalf("failed connect to database: %v", err)
	}
	defer db.Close()

	apiConfig := handlers.NewConfig(db, tokenSecret)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", apiConfig.Register)
	mux.HandleFunc("POST /login", apiConfig.Login)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Println("Staring server on port 8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
