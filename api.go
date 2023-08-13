package main

import (
	"encoding/json"
	"net/http"

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

//
// func (s *APIServer) handleGetTodos(w http.ResponseWriter, r *http.Request) error {
// 	rows, err := s.store.GetTodos()
// 	if err != nil {
// 		return err
// 	}
// 	w.Write([]byte(fmt.Sprint(rows)))
// 	return nil
// }
