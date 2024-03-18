package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		r.Patch("/", handleError(s.handlePatchTodo))
	})
	router.Get("/tags/{tag}", handleError(s.handleGetTodosByTag))
	router.Get("/tags", handleError(s.handleGetTags))
	router.Post("/tags", handleError(s.handleCreateTag))
	http.ListenAndServe(s.listenAddr, router)
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error
type APIError struct {
	Text   string
	Status int // necessary?
}

func handleError(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, APIError{fmt.Sprintf("API error: %v/%s\n", err, err.Error()), http.StatusBadRequest})
		}
	}
}

func (s *APIServer) handleGetTodos(w http.ResponseWriter, r *http.Request) error {
	// TODO min fallback 0 max cap/fallback 10 (as in the db)
	tq, err := queryFromRequest(r)
	q, _ := tq.SQL()
	log.Printf("Created query:\n%v\n", q)
	rows, err := s.store.QueryTodos(tq)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, rows)
}

func queryFromRequest(r *http.Request) (*TodoQuery, error) {
	q := r.URL.Query()
	opt := make([]TodoQueryOption, 0)
	if q.Has("completed") {
		completed := false
		switch val := q.Get("completed"); val {
		case "true":
			completed = true
		case "false":
		default:
			return nil, fmt.Errorf("Invalid bool format: %s\n", val)
		}
		opt = append(opt, WithCompleted(completed))
	}
	// TODO: validation here, but could be made in the query struct
	//       maybe custom validation for tagged fields?
	//       - int range etc.
	if q.Has("prioritymin") {
		prioritymin, err := strconv.Atoi(q.Get("prioritymin"))
		if err != nil {
			return nil, fmt.Errorf("Invalid int format: %s", q.Get("prioritymin"))
		}
		opt = append(opt, WithPriorityMin(prioritymin))
	}
	if q.Has("prioritymax") {
		prioritymax, err := strconv.Atoi(q.Get("prioritymax"))
		if err != nil {
			return nil, fmt.Errorf("Invalid int format: %s", q.Get("prioritymax"))
		}
		opt = append(opt, WithPriorityMax(prioritymax))
	}
	if q.Has("fulltext") {
		val := q.Get("fulltext")
		if len(val) < 3 {
			return nil, fmt.Errorf("Query too short, at least 3 characters required.")
		}
		opt = append(opt, WithFulltext(val))
	}

	tq := NewTodoQuery(opt)
	return &tq, nil
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

func (s *APIServer) handlePatchTodo(w http.ResponseWriter, r *http.Request) error {
	id, _ := r.Context().Value("id").(int)

	var item = Todo{
		Id:        id,
		Completed: chi.URLParam(r, "completed") != "",
	}

	patchOptions := PatchOptions{Completed: true}
	err := s.store.PatchTodo(&item, patchOptions)

	return err
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}
