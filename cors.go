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
	if origin == "" {
		return
	}

	if len(s.CorsOptions.Origins) > 0 && s.CorsOptions.Origins[0] == "*" {
		if s.CorsOptions.AllowCredentials {
			res.Headers.Add("Access-Control-Allow-Origin", origin)
		} else {
			res.Headers.Add("Access-Control-Allow-Origin", "*")
		}
	} else if contains(s.CorsOptions.Origins, origin) {
		res.Headers.Add("Access-Control-Allow-Origin", origin)
	} else {
		// Origin not allowed
		return
	}

	if s.CorsOptions.AllowCredentials {
		res.Headers.Add("Access-Control-Allow-Credentials", "true")
	}

	if req.Method == "OPTIONS" {
		if len(s.CorsOptions.AllowedMethods) > 0 {
			if s.CorsOptions.AllowedMethods[0] == "*" {
				res.Headers.Add("Access-Control-Allow-Methods",
					"GET,POST,PUT,DELETE,OPTIONS,HEAD,PATCH")
			} else {
				res.Headers.Add("Access-Control-Allow-Methods",
					strings.Join(s.CorsOptions.AllowedMethods, ","))
			}
		}

		if len(s.CorsOptions.AllowedHeaders) > 0 {
			if s.CorsOptions.AllowedHeaders[0] == "*" {
				if reqHeaders, ok := req.Headers["access-control-request-headers"]; ok && reqHeaders != "" {
					res.Headers.Add("Access-Control-Allow-Headers", reqHeaders)
				} else {
					res.Headers.Add("Access-Control-Allow-Headers",
						"Content-Type,Authorization,Accept,Origin,X-Requested-With")
				}
			} else {
				res.Headers.Add("Access-Control-Allow-Headers",
					strings.Join(s.CorsOptions.AllowedHeaders, ","))
			}
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
