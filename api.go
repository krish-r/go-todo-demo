package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	todoPattern     = "/todo/"
	dateTimePattern = "2006-01-02 03:04:05PM"
)

const (
	emptyId   = -1
	invalidId = -2
)

var (
	ch chan int
)

type server struct {
	addr  string
	store storage
}

func NewServer(addr string, store storage) *server {
	return &server{
		addr:  addr,
		store: store,
	}
}

func (s *server) start() {
	ch = make(chan int)
	go generateID(ch)

	http.HandleFunc(todoPattern, todoAPI(s.handleTodo).decorate())

	fmt.Print(http.ListenAndServe(s.addr, nil))
	os.Exit(1)
}

func (s *server) handleTodo(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetTodo(w, r)
	case "POST":
		return s.handleAddTodo(w, r)
	case "DELETE":
		return s.handleDeleteTodo(w, r)
	default:
		return fmt.Errorf("Invalid Method %q", r.Method)
	}
}

func (s *server) handleGetTodo(w http.ResponseWriter, r *http.Request) error {
	url := r.URL.Path
	id, err := parseID(url)
	if err != nil {
		return err
	}

	if id == emptyId {
		return s.handleGetAllTodo(w, r)
	}

	t, err := s.store.get(id)
	if err != nil {
		return err
	}

	err = writeJSON(w, http.StatusOK, *t)
	return err
}

func (s *server) handleGetAllTodo(w http.ResponseWriter, r *http.Request) error {
	t, err := s.store.getAll()
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, t)
}

func (s *server) handleAddTodo(w http.ResponseWriter, r *http.Request) error {
	t := &addTodo{}
	var err error
	if err = readJSON(r.Body, t); err != nil {
		return err
	}

	due, err := time.Parse(dateTimePattern, t.Due)
	if err != nil {
		return err
	}

	id, err := findNextAvailableID(s)
	if err != nil {
		return err
	}
	if id == invalidId {
		return fmt.Errorf("Unable to generate an unique id")
	}

	newTodo := NewTodo(id, t.Description, due)

	if err = s.store.add(newTodo); err != nil {
		return err
	}

	if err = writeJSON(w, http.StatusCreated, *newTodo); err != nil {
		return err
	}
	err = r.Body.Close()
	return err
}

func (s *server) handleDeleteTodo(w http.ResponseWriter, r *http.Request) error {
	url := r.URL.Path
	id, err := parseID(url)
	if err != nil {
		return err
	}

	if id == emptyId {
		return fmt.Errorf("id cannot be empty")
	}

	found, err := s.store.contains(id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("id not found")
	}

	err = s.store.delete(id)
	return err
}

func writeJSON[V todo | todoError | []*todo](w http.ResponseWriter, statusCode int, data V) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	return err
}

func readJSON(r io.Reader, t *addTodo) error {
	err := json.NewDecoder(r).Decode(t)
	return err
}

type todoAPI func(w http.ResponseWriter, r *http.Request) error

func (t todoAPI) decorate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := t(w, r); err != nil {
			err := writeJSON(w, http.StatusBadRequest, todoError{Error: err.Error()})
			check(err)
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func generateID(ch chan<- int) {
	i := 0
	for {
		i += 1
		ch <- i
	}
}

func findNextAvailableID(s *server) (int, error) {
	var id int
	for {
		id = <-ch
		exists, err := s.store.contains(id)
		if err != nil {
			return invalidId, err
		}
		if !exists {
			break
		}
	}
	return id, nil
}

func parseID(url string) (int, error) {
	id := strings.TrimSpace(strings.Replace(url, todoPattern, "", 1))
	if isEmpty(id) {
		return emptyId, nil
	}
	return strconv.Atoi(id)
}

func isEmpty(s string) bool {
	return s == ""
}
