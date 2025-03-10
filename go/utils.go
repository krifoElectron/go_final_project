package main

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func NextDate(nowArg time.Time, eventDateString string, repeat string) (string, error) {
	now := time.Date(nowArg.Year(), nowArg.Month(), nowArg.Day(), 0, 0, 0, nowArg.Nanosecond(), nowArg.Location())
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

			nextDate := eventDate
			for nextDate.Before(now) {
				nextDate = nextDate.AddDate(0, 0, period)
			}
			return nextDate.Format("20060102"), nil
		}
	}

	err = errors.New("неверный формат repeat")
	return "", err
}
