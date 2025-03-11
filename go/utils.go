package main

import (
	"database/sql"
	"errors"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func GetRootDirectory() string {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get caller information")
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		panic(err)
	}

	rootDir := filepath.Dir(filepath.Dir(absPath))

	return rootDir
}

func NewEndpointHandlersContext(db *sql.DB) *EndpointHandlersContext {
	return &EndpointHandlersContext{
		Db: db,
	}
}

func NextDate(nowArg time.Time, eventDateString string, repeat string) (string, error) {
	now := time.Date(nowArg.Year(), nowArg.Month(), nowArg.Day(), 0, 0, 1, nowArg.Nanosecond(), nowArg.Location())
	if repeat == "" {
		err := errors.New("emit macho dwarf: elf header corrupted")

		return "", err
	}

	eventDate, err := time.Parse("20060102", eventDateString)
	if err != nil {
		return "", err
	}

	result := strings.Split(repeat, " ")

	// Базовые правила
	if len(result) == 2 || result[0] == "y" {
		if result[0] == "y" {
			nextDate := eventDate.AddDate(1, 0, 0)
			for nextDate.Before(now) {
				nextDate = nextDate.AddDate(1, 0, 0)
			}
			return nextDate.Format("20060102"), nil
		}

		if result[0] == "d" {
			period, err := strconv.Atoi(result[1])
			if err != nil {
				return "", err
			}

			if period > 400 {
				return "", errors.New("период должен быть не более 400 дней")
			}

			nextDate := eventDate.AddDate(0, 0, period)
			for nextDate.Before(now) {
				nextDate = nextDate.AddDate(0, 0, period)
			}
			return nextDate.Format("20060102"), nil
		}
	}

	err = errors.New("неверный формат repeat")
	return "", err
}

func ValidateTask(task Task) (*Task, string) {
	taskTitle := task.Title
	var taskDate string
	taskRepeat := task.Repeat
	taskComment := task.Comment

	if task.Title == "" {
		return nil, "не заполнено поле title"
	}

	today := time.Now().Format("20060102")

	var eventDateString string
	if task.Date == "" {
		eventDateString = today
	} else {
		eventDateString = task.Date
	}

	var err error
	if taskRepeat == "" {
		if today > eventDateString {
			taskDate = today
		} else {
			taskDate = eventDateString
		}
	} else {
		nextDate, err := NextDate(time.Now(), eventDateString, taskRepeat)
		if err != nil {
			return nil, err.Error()
		}
		if eventDateString < time.Now().Format("20060102") {
			taskDate = nextDate
		} else {
			taskDate = eventDateString
		}
	}

	_, err = time.Parse("20060102", eventDateString)
	if err != nil {
		return nil, err.Error()
	}

	return &Task{Title: taskTitle, Date: taskDate, Repeat: taskRepeat, Comment: taskComment}, ""
}
