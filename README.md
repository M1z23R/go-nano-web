# go-nano-web

GoNanoWeb is a minimal/lightweight HTTP web framework written in Go (Golang).

Currently, it supports the following features:

- Routers
- Middlewares
- CORS
- JSON Body parser out of the box
- FormData handling with streaming capabilities
- Rate limiting
- Security features (CSRF protection, security headers)
- Passing data down the chain

Example:

```go
package main

import (
	"fmt"
	"time"

	gonanoweb "github.com/M1z23R/go-nano-web"
)

func main() {
	limiter1 := gonanoweb.NewRateLimiter(5, 5*time.Second)
	limiter2 := gonanoweb.NewRateLimiter(5, 5*time.Second)
	defer limiter1.Stop()
	defer limiter2.Stop()

	corsOptions := CorsOptions()
	formDataOptions := gonanoweb.DefaultFormDataOptions()
	formDataOptions.StreamingParser = true

	g := gonanoweb.NewServer(":1337", &gonanoweb.ServerOptions{
		CorsOptions: &corsOptions,
	})
	
	// Basic endpoint
	g.Get("/api", get)
	
	// With rate limiting
	g.Get("/api/limited", get, gonanoweb.RateLimitMiddleware(limiter1))
	
	// With form data handling
	g.Post("/api/upload", handleFormData, gonanoweb.FormDataMiddleware(&formDataOptions))
	
	// Global middleware
	g.UseMiddleware(gonanoweb.RateLimitMiddleware(limiter2))
	
	g.Listen()
}

func get(res *gonanoweb.Response, req *gonanoweb.Request) error {
	var resp = map[string]string{}
	resp["message"] = "Hello world!"

	res.Json(200, resp)
	return nil
}

// Handler for form data
func handleFormData(res *gonanoweb.Response, req *gonanoweb.Request) error {
	formData, err := gonanoweb.GetFormData(req)
	if err != nil {
		return err
	}

	response := map[string]interface{}{
		"message": "Upload successful",
		"fields":  formData.Fields,
		"files":   []map[string]string{},
	}

	// Process files
	for name, files := range formData.Files {
		for _, file := range files {
			fileInfo := map[string]string{
				"fieldName": name,
				"filename":  file.Filename,
				"size":      fmt.Sprintf("%d bytes", file.Size),
			}
			response["files"] = append(response["files"].([]map[string]string), fileInfo)
		}
	}

	res.Json(200, response)
	return nil
}

// Streaming example would access individual parts as they arrive
func handleStreamingFormData(res *gonanoweb.Response, req *gonanoweb.Request) error {
	reader, err := gonanoweb.GetMultipartReader(req)
	if err != nil {
		return err
	}
	
	formData, _ := gonanoweb.GetFormData(req)
	
	// Process parts as they arrive
	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}
		
		name := part.FormName()
		filename := part.FileName()
		
		if filename != "" {
			// Handle file part
			// Process stream directly without loading whole file to memory
		} else {
			// Handle regular field
			value, _ := io.ReadAll(part)
			formData.Fields[name] = append(formData.Fields[name], string(value))
		}
	}
	
	res.Json(200, map[string]string{"message": "Streaming upload processed"})
	return nil
}

func CorsOptions() gonanoweb.CorsOptions {
	return gonanoweb.CorsOptions{
		Origins:          []string{"https://en.wikipedia.org"},
		AllowedMethods:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{},
		AllowCredentials: true,
		MaxAge:           3600,
	}
}
```
