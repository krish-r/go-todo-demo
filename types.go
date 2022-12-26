package main

import (
	"time"
)

type todo struct {
	Id          int       `json:"id" bson:"id"`
	Description string    `json:"description" bson:"description"`
	Created     time.Time `json:"created" bson:"created"`
	Due         time.Time `json:"due" bson:"due"`
}

func NewTodo(id int, description string, due time.Time) *todo {
	return &todo{
		Id:          id,
		Description: description,
		Created:     time.Now().UTC(),
		Due:         due.UTC(),
	}
}

type addTodo struct {
	Description string `json:"description"`
	Due         string `json:"due"`
}

type todoError struct {
	Error string `json:"error"`
}

func NewTodoError(err string) *todoError {
	return &todoError{
		Error: err,
	}
}
