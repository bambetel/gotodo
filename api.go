package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	chi "github.com/go-chi/chi/v5"
)

type APIServer struct {
	listenAddr string
	store      *postgresStorage
}

func NewAPIServer(listenAddr string, store *postgresStorage) *APIServer {
	return &APIServer{listenAddr: listenAddr, store: store}
}

func (s *APIServer) Run() {
	router := chi.NewRouter()
	router.Get("/todos", s.handleGetTodos)
	router.Post("/todos", s.handleCreateTodo)
	router.Delete("/todos", s.handleDeleteTodo)
	router.Put("/todos/{id}", s.handleUpdateTodo)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleGetTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.GetTodos()
	if err != nil {
		w.Write([]byte("error!!!"))
	} else {
		json.NewEncoder(w).Encode(rows)
	}
}

func (s *APIServer) handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	var (
		label string = r.FormValue("label")
	)
	priority, _ := strconv.Atoi(r.FormValue("priority"))

	var ins = Todo{
		Label:    label,
		Priority: priority,
	}
	item, err := s.store.CreateTodo(&ins)
	if err != nil {
		w.Write([]byte("error creating todo!"))
	} else {
		json.NewEncoder(w).Encode(item)
	}

}

func (s *APIServer) handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		w.Write([]byte("Cannot parse id from form data!"))
	}
	err = s.store.DeleteTodo(id)
	if err != nil {
		w.Write([]byte("Error deleting item"))
	} else {
		w.Write([]byte("Item deleted."))
	}
}

func (s *APIServer) handleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		w.Write([]byte("Cannot parse id from chi url!"))
	}
	priority, err := strconv.Atoi(r.FormValue("priority"))

	if err != nil {
		w.Write([]byte("Cannot parse priority from form data!"))
	}
	label := r.FormValue("label")
	completed := r.FormValue("completed") == "true" // bool

	// TODO handling partial updates
	var item = Todo{
		Id:        id,
		Label:     label,
		Priority:  priority,
		Completed: completed,
	}
	val, err := s.store.UpdateTodo(&item)
	if err != nil {
		w.Write([]byte("Error updating item"))
	} else {
		w.Write([]byte(fmt.Sprint("Item updated %+v", val)))
	}
}

//
// func (s *APIServer) handleGetTodos(w http.ResponseWriter, r *http.Request) error {
// 	rows, err := s.store.GetTodos()
// 	if err != nil {
// 		return err
// 	}
// 	w.Write([]byte(fmt.Sprint(rows)))
// 	return nil
// }
