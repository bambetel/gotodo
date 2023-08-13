package main

import (
	"fmt"
	"log"
)

func main() {
	ps, err := NewPostgresStorage()
	if err != nil {
		log.Fatal("error creating pg connection!")
	}

	todos, err := ps.GetTodos()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(todos)

	server := NewAPIServer(":3000", ps)
	server.Run()
}
