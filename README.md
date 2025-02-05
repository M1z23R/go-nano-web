# go-nano-web

GoNanoWeb is a minimal/lightweight HTTP web framework written in Go (Golang).

Currently, it supports the following features:

- Routers
- Middlewares
- Cors
- Json Body parser out of the box
- Passing data down the chain

Example:

```
package main

import (
	"time"

	gonanoweb "github.com/M1z23R/go-nano-web"
)

func main() {

	limiter1 := gonanoweb.NewRateLimiter(5, 5*time.Second)
	limiter2 := gonanoweb.NewRateLimiter(5, 5*time.Second)
	defer limiter1.Stop()
	defer limiter2.Stop()

	corsOptions := CorsOptions()
	g := gonanoweb.NewServer(":1337", &gonanoweb.ServerOptions{
		CorsOptions: &corsOptions,
	})
	g.Get("/1", get)
	g.Get("/2", get, gonanoweb.RateLimitMiddleware(limiter1))
	g.UseMiddleware(gonanoweb.RateLimitMiddleware(limiter2))
	g.Get("/3", get)
	g.Listen()

}

func get(res *gonanoweb.Response, req *gonanoweb.Request) error {
	var resp = map[string]string{}
	resp["message"] = "Hello world!"

	res.Json(200, resp)

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
