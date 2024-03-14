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
	router.Get("/todos", s.handleGetTodos)
	router.Post("/todos", s.handleCreateTodo)
	router.Route("/todos/{id}", func(r chi.Router) {
		r.Use(TodoCtx)
		r.Delete("/", s.handleDeleteTodo)
		r.Post("/tags", s.handleAddItemTag)
		r.Delete("/tags/{label}", s.handleRmItemTag)
		r.Put("/", s.handleUpdateTodo)
	})
	router.Get("/tags/{tag}", s.handleGetTodosByTag)
	router.Get("/tags", s.handleGetTags)
	router.Post("/tags", s.handleCreateTag)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleGetTodos(w http.ResponseWriter, r *http.Request) {
	// TODO min fallback 0 max cap/fallback 10 (as in the db)
	q := r.URL.Query()
	var (
		minPriority = q.Get("minpriority")
		maxPriority = q.Get("maxpriority")
		// orderBy     = q.Get("sort")
		// orderDir    = q.Get("dir")
		// completed   = q.Get("completed")
		// fulltext    = q.Get("q")
	)
	fmt.Printf("url query priority range %v..%v\n", minPriority, maxPriority)
	rows, err := s.store.GetTodos()
	if err != nil {
		w.Write([]byte("error!!!" + err.Error()))
	} else {
		WriteJSON(w, http.StatusOK, rows)
	}
}

func (s *APIServer) handleGetTodosByTag(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")
	fmt.Println("tag:", tag)
	rows, err := s.store.GetTodosByTag(tag)
	if err != nil {
		w.Write([]byte("error!!!" + err.Error()))
	} else {
		WriteJSON(w, http.StatusOK, rows)
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

func (s *APIServer) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	label := r.FormValue("label")
	if err := s.store.CreateTag(label); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Error creating tag: %v\n", err.Error())))
	}
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

func (s *APIServer) handleAddItemTag(w http.ResponseWriter, r *http.Request) {
	id, _ := r.Context().Value("id").(int)
	label := r.FormValue("label")
	if err := s.store.AddItemTag(id, label); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *APIServer) handleRmItemTag(w http.ResponseWriter, r *http.Request) {
	id, _ := r.Context().Value("id").(int)
	label := chi.URLParam(r, "label")
	if err := s.store.RmItemTag(id, label); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *APIServer) handleGetTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.store.GetTags()
	if err != nil {
		w.Write([]byte("error getting tag list"))
	}
	json.NewEncoder(w).Encode(tags)
}

func (s *APIServer) handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, _ := r.Context().Value("id").(int)
	err := s.store.DeleteTodo(id)
	if err != nil {
		w.Write([]byte("Error deleting item"))
	} else {
		w.Write([]byte("Item deleted."))
	}
}

func (s *APIServer) handleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, _ := r.Context().Value("id").(int)
	priority, err := strconv.Atoi(r.FormValue("priority"))

	if err != nil {
		w.Write([]byte("Cannot parse priority from form data!"))
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
	val, err := s.store.UpdateTodo(&item)
	if err != nil {
		w.Write([]byte("Error updating item"))
	} else {
		w.Write([]byte(fmt.Sprintf("Item updated %+v", val)))
	}
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// TODO ?
// func ErrorJSON(w http.ResponseWriter, status int, data any) error {
//    ... return JSON with error message
// }
