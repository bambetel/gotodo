package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	connStr := fmt.Sprintf("user=%s password=%s dbname=todo host=/tmp sslmode=disable", os.Getenv("PGUSER"), os.Getenv("PGPASS"))
	ps, err := NewPostgresStorage(connStr)
	if err != nil {
		log.Fatal("error creating pg connection!")
	}

	server := NewAPIServer(":3000", ps)
	server.Run()
}
