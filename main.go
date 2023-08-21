package main

import (
	"log"
)

func main() {
	ps, err := NewPostgresStorage()
	if err != nil {
		log.Fatal("error creating pg connection!")
	}

	server := NewAPIServer(":3000", ps)
	server.Run()
}
