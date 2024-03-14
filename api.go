package main

import (
	"context"
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
	router.Get("/todos", handleError(s.handleGetTodos))
	router.Post("/todos", handleError(s.handleCreateTodo))
	router.Route("/todos/{id}", func(r chi.Router) {
		r.Use(TodoCtx)
		r.Delete("/", handleError(s.handleDeleteTodo))
		r.Post("/tags", handleError(s.handleAddItemTag))
		r.Delete("/tags/{label}", handleError(s.handleRmItemTag))
		r.Put("/", handleError(s.handleUpdateTodo))
	})
	router.Get("/tags/{tag}", handleError(s.handleGetTodosByTag))
	router.Get("/tags", handleError(s.handleGetTags))
	router.Post("/tags", handleError(s.handleCreateTag))
	http.ListenAndServe(s.listenAddr, router)
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error

func handleError(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("API error: %v/%s\n", err, err.Error()))
		}
	}
}

func (s *APIServer) handleGetTodos(w http.ResponseWriter, r *http.Request) error {
	// TODO min fallback 0 max cap/fallback 10 (as in the db)
	q := r.URL.Query()
	var (
		minPriority = q.Get("minpriority")
		maxPriority = q.Get("maxpriority")
		// orderBy     = q.Get("sort")
		// orderDir    = q.Get("dir")
		completed = q.Get("completed")
		fulltext  = q.Get("q")
	)
	fmt.Printf("url query priority range %v..%v, %v, %v\n", minPriority, maxPriority, completed, fulltext)
	rows, err := s.store.GetTodos()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, rows)
}

func (s *APIServer) handleGetTodosByTag(w http.ResponseWriter, r *http.Request) error {
	tag := chi.URLParam(r, "tag")
	fmt.Println("tag:", tag)
	rows, err := s.store.GetTodosByTag(tag)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, rows)
}

func (s *APIServer) handleCreateTodo(w http.ResponseWriter, r *http.Request) error {
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
		return err
	}
	return WriteJSON(w, http.StatusCreated, item)
}

func (s *APIServer) handleCreateTag(w http.ResponseWriter, r *http.Request) error {
	label := r.FormValue("label")
	return s.store.CreateTag(label)
}

func TodoCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Cannot parse id from query!"))
			return
		}
		ctx := context.WithValue(r.Context(), "id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *APIServer) handleAddItemTag(w http.ResponseWriter, r *http.Request) error {
	id, _ := r.Context().Value("id").(int)
	label := r.FormValue("label")
	return s.store.AddItemTag(id, label)
}

func (s *APIServer) handleRmItemTag(w http.ResponseWriter, r *http.Request) error {
	id, _ := r.Context().Value("id").(int)
	label := chi.URLParam(r, "label")
	return s.store.RmItemTag(id, label)
}

func (s *APIServer) handleGetTags(w http.ResponseWriter, r *http.Request) error {
	tags, err := s.store.GetTags()
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(tags)
}

func (s *APIServer) handleDeleteTodo(w http.ResponseWriter, r *http.Request) error {
	id, _ := r.Context().Value("id").(int)
	return s.store.DeleteTodo(id)
}

func (s *APIServer) handleUpdateTodo(w http.ResponseWriter, r *http.Request) error {
	id, _ := r.Context().Value("id").(int)
	priority, err := strconv.Atoi(r.FormValue("priority"))

	if err != nil {
		return err
	}
	label := r.FormValue("label")
	completed := r.FormValue("completed") == "true" // bool
	tags := r.FormValue("tags")

	// TODO handling partial updates
	var item = Todo{
		Id:        id,
		Label:     label,
		Priority:  priority,
		Completed: completed,
		Tags:      tags,
	}
	_, err = s.store.UpdateTodo(&item)
	return err
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}
