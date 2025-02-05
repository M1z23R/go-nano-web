package gonanoweb

import (
	"log"
	"strings"
	"time"
)

func validatePath(path string) {
	if !strings.HasPrefix(path, "/") {
		log.Panic("path must start with /")
	}
	if strings.Contains(path, "..") {
		log.Panic("path cannot contain ..")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func currentDateString() string {
	time := time.Now().UTC().Format(time.RFC1123)
	return "Date: " + time[:len(time)-3] + "GMT"
}

func removeEmpty(v string) bool {
	return v == ""
}
