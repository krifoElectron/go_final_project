package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

func returnError(w http.ResponseWriter, stringErr string) {
	errResp := ErrorResponse{Error: stringErr}
	errRespByte, _ := json.Marshal(errResp)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(errRespByte)
}

func NedxDateEndpoint(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	nowDate, err := time.Parse("20060102", now)
	if err != nil {
		w.Write([]byte("Неверный формат now"))
		return
	}

	nextDate, err := NextDate(nowDate, date, repeat)

	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte(nextDate))
}

func AddTaskEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)

	if err != nil {
		returnError(w, err.Error())
		return
	}

	var task Task
	err = json.Unmarshal(body, &task)
	if err != nil {
		returnError(w, err.Error())
		return
	}

	taskTitle := task.Title
	var taskDate string
	taskRepeat := task.Repeat
	taskComment := task.Comment

	if task.Title == "" {
		returnError(w, "не заполнено поле title")
		return
	}

	today := time.Now().Format("20060102")

	var eventDateString string
	if task.Date == "" {
		eventDateString = today
	} else {
		eventDateString = task.Date
	}

	if taskRepeat == "" {
		if today > eventDateString {
			taskDate = today
		} else {
			taskDate = eventDateString
		}
	} else {
		taskDate, err = NextDate(time.Now(), eventDateString, taskRepeat)
		if err != nil {
			returnError(w, err.Error())
			return
		}
	}

	_, err = time.Parse("20060102", eventDateString)
	if err != nil {
		returnError(w, err.Error())
		return
	}

	db := GetDB()
	sqlResult, err := db.Exec(
		`INSERT INTO scheduler (date, repeat, title, comment) VALUES (?, ?, ?, ?)`,
		taskDate, taskRepeat, taskTitle, taskComment,
	)
	if err != nil {
		returnError(w, err.Error())
	}

	id, err := sqlResult.LastInsertId()
	if err != nil {
		returnError(w, err.Error())
	}

	var response TaskResponse = TaskResponse{
		Id: strconv.FormatInt(id, 10),
	}

	w.WriteHeader(http.StatusOK)
	jsonResponse, _ := json.Marshal(response)
	w.Write(jsonResponse)
}
