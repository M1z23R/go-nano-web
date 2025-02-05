package gonanoweb

import (
	"strconv"
	"strings"
)

type CorsOptions struct {
	Origins          []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func (s *Server) handleCORS(res *Response, req *Request) {
	if s.CorsOptions == nil {
		return
	}

	origin := req.Headers["origin"]
	if !contains(s.CorsOptions.Origins, origin) && s.CorsOptions.Origins[0] != "*" {
		return
	}

	res.Headers.Add("Access-Control-Allow-Origin", origin)

	if s.CorsOptions.AllowCredentials {
		res.Headers.Add("Access-Control-Allow-Credentials", "true")
	}

	if req.Method == "OPTIONS" {
		if len(s.CorsOptions.AllowedMethods) > 0 {
			res.Headers.Add("Access-Control-Allow-Methods",
				strings.Join(s.CorsOptions.AllowedMethods, ","))
		}
		if len(s.CorsOptions.AllowedHeaders) > 0 {
			res.Headers.Add("Access-Control-Allow-Headers",
				strings.Join(s.CorsOptions.AllowedHeaders, ","))
		}
		if s.CorsOptions.MaxAge > 0 {
			res.Headers.Add("Access-Control-Max-Age",
				strconv.Itoa(s.CorsOptions.MaxAge))
		}
	}

	if len(s.CorsOptions.ExposedHeaders) > 0 {
		res.Headers.Add("Access-Control-Expose-Headers",
			strings.Join(s.CorsOptions.ExposedHeaders, ","))
	}
}
