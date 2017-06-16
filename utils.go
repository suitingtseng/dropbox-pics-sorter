package main

import (
	"fmt"
	"strings"
)

type DirSet map[string]bool

func (s DirSet) Add(arg MkdirArg) {
	key := fmt.Sprintf("%s/%s", arg.base, arg.date.Format(FORMAT))
	s[key] = true
}

func (s DirSet) Contains(arg MkdirArg) bool {
	key := fmt.Sprintf("%s/%s", arg.base, arg.date.Format(FORMAT))
	return s[key]
}

func isImage(path string) bool {
	formats := []string{".jpg", ".png", ".gif"}
	for _, format := range formats {
		if strings.HasSuffix(path, format) {
			return true
		}
	}
	return false
}

func isVideo(path string) bool {
	formats := []string{".mov", ".mp4"}
	for _, format := range formats {
		if strings.HasSuffix(path, format) {
			return true
		}
	}
	return false
}
