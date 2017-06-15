package main

import (
	"strings"
	"time"
)

type DateSet map[string]bool

func (s DateSet) Add(t time.Time) {
	s[t.Format(FORMAT)] = true
}

func (s DateSet) Contains(t time.Time) bool {
	return s[t.Format(FORMAT)]
}

func isImage(path string) bool {
	formats := []string{".jpg", ".png"}
	for _, format := range formats {
		if strings.HasSuffix(path, format) {
			return true
		}
	}
	return false
}
