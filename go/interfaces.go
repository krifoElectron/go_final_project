package main

import "database/sql"

type EndpointHandlersContext struct {
	Db *sql.DB
}

type Task struct {
	Date    string `json:"date,omitempty"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat"`
}

type TaskResponse struct {
	Id string `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TaskWithId struct {
	Task
	Id string `json:"id"`
}

type TasksResponse struct {
	Tasks []TaskWithId `json:"tasks"`
}
