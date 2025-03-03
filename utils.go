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
	
	// Check for path traversal attempts, including URL encoded versions
	if strings.Contains(path, "..") || 
	   strings.Contains(path, "%2e%2e") || 
	   strings.Contains(path, "%2E%2E") {
		log.Panic("path cannot contain path traversal sequences")
	}
	
	// Prevent control characters and null bytes in paths
	for _, r := range path {
		if r < 32 || r == 127 {
			log.Panic("path contains invalid characters")
		}
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
