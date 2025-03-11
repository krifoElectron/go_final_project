package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func RegisterRoutes(handlerContext *EndpointHandlersContext) {
	rootDir := GetRootDirectory()
	fmt.Println("Для проверки rootDir:", rootDir)

	http.Handle("/", http.FileServer(http.Dir(rootDir+"/web")))
	http.Handle("/api/nextdate", http.HandlerFunc(NextDateEndpoint))
	http.Handle("/api/task", http.HandlerFunc(handlerContext.TaskEndpoint))
	http.Handle("/api/tasks", http.HandlerFunc(handlerContext.GetTasksEndpoint))
	http.Handle("/api/task/done", http.HandlerFunc(handlerContext.DoneEndpoint))
}

func returnError(w http.ResponseWriter, stringErr string) {
	errResp := ErrorResponse{Error: stringErr}
	errRespByte, _ := json.Marshal(errResp)
	w.Write(errRespByte)
}

func NextDateEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	nowDate, err := time.Parse("20060102", now)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, "Неверный формат now")
		return
	}

	nextDate, err := NextDate(nowDate, date, repeat)

	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte(nextDate))
}

func (hCtx *EndpointHandlersContext) TaskEndpoint(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		hCtx.AddTaskEndpoint(w, r)
	case "GET":
		hCtx.GetTaskEndpoint(w, r)
	case "PUT":
		hCtx.UpdateTask(w, r)
	case "DELETE":
		hCtx.DeleteTaskEndpoint(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (hCtx *EndpointHandlersContext) AddTaskEndpoint(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, err.Error())
		return
	}

	var task Task
	err = json.Unmarshal(body, &task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, err.Error())
		return
	}

	taskToSave, strErr := ValidateTask(task)
	if strErr != "" {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, strErr)
		return
	}

	sqlResult, err := hCtx.Db.Exec(
		`INSERT INTO scheduler (date, repeat, title, comment) VALUES (?, ?, ?, ?)`,
		taskToSave.Date, taskToSave.Repeat, taskToSave.Title, taskToSave.Comment,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}

	id, err := sqlResult.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}

	var response TaskResponse = TaskResponse{
		Id: strconv.FormatInt(id, 10),
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (hCtx *EndpointHandlersContext) GetTaskEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.FormValue("id")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, "Не указан идентификатор")
		return
	}

	rows, err := hCtx.Db.Query(`SELECT id, title, date, repeat, comment
		FROM scheduler
		WHERE id=?;`, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}

	defer rows.Close()
	var resp []byte
	ok := rows.Next()
	if ok {
		var task TaskWithId
		if err := rows.Scan(&task.Id, &task.Title, &task.Date, &task.Repeat, &task.Comment); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			returnError(w, err.Error())
			return
		}

		resp, err = json.Marshal(task)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			returnError(w, err.Error())
			return
		}

		err = rows.Err()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			returnError(w, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	returnError(w, "Задача не найдена")
}

func (hCtx *EndpointHandlersContext) UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, err.Error())
		return
	}

	var taskWithId TaskWithId
	err = json.Unmarshal(body, &taskWithId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, err.Error())
		return
	}

	task := Task{Title: taskWithId.Title, Date: taskWithId.Date, Repeat: taskWithId.Repeat, Comment: taskWithId.Comment}
	id := taskWithId.Id
	taskToSave, strErr := ValidateTask(task)
	if strErr != "" {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, strErr)
		return
	}

	updateQuery := `
		UPDATE scheduler
		SET date = ?, repeat = ?, title = ?, comment = ?
		WHERE id = ?`

	result, err := hCtx.Db.Exec(
		updateQuery,
		taskToSave.Date, taskToSave.Repeat, taskToSave.Title, taskToSave.Comment, id,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, "Задача не найдена")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (hCtx *EndpointHandlersContext) GetTasksEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var response TasksResponse

	rows, err := hCtx.Db.Query(`SELECT id, title, date, repeat, comment
		FROM scheduler
		ORDER BY date ASC
		LIMIT 20;`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}

	defer rows.Close()
	for rows.Next() {
		var task TaskWithId
		if err := rows.Scan(&task.Id, &task.Title, &task.Date, &task.Repeat, &task.Comment); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			returnError(w, err.Error())
			return
		}
		response.Tasks = append(response.Tasks, task)
	}

	err = rows.Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}

	var resp []byte
	if response.Tasks == nil {
		resp = []byte("{ \"tasks\": [] }")
	} else {
		resp, err = json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			returnError(w, err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (hCtx *EndpointHandlersContext) DoneEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, "не указан идентификатор")
		return
	}

	rows, err := hCtx.Db.Query(`SELECT id, title, date, repeat, comment
		FROM scheduler
		WHERE id=?;`, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}

	var task TaskWithId

	ok := rows.Next()
	if ok {
		if err := rows.Scan(&task.Id, &task.Title, &task.Date, &task.Repeat, &task.Comment); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			returnError(w, err.Error())
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, "Задача не найдена")
		return
	}
	err = rows.Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}
	rows.Close()

	w.Header().Set("Content-Type", "application/json")

	if task.Repeat == "" {
		doneQuery := `
			DELETE FROM scheduler
			WHERE id=?`

		_, err := hCtx.Db.Exec(doneQuery, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			returnError(w, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}

	updateQuery := `
		UPDATE scheduler
		SET date = ?
		WHERE id = ?`

	// считаем, что тут не может быть ошибки, поскольку данные в БД уже валидны
	nextDate, _ := NextDate(time.Now(), task.Date, task.Repeat)

	_, err = hCtx.Db.Exec(
		updateQuery,
		nextDate,
		id,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (hCtx *EndpointHandlersContext) DeleteTaskEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, "не указан идентификатор")
		return
	}

	deleteQuery := `
		DELETE FROM scheduler
		WHERE id=?`

	sqlResult, err := hCtx.Db.Exec(deleteQuery, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}

	rowsAffected, err := sqlResult.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		returnError(w, err.Error())
		return
	}
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusBadRequest)
		returnError(w, "Задача не найдена")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
