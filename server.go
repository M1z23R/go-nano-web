package gonanoweb

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"
)

func NewServer(addr string) *Server {
	return &Server{
		EventStreams: make(map[string]*chan string),
		Stack:        []IStackable{},
		addr:         addr,
	}
}

type Server struct {
	addr         string
	listener     net.Listener
	Stack        []IStackable
	EventStreams map[string]*chan string
	CorsOptions  *CorsOptions
}

func (s *Server) GetStack() []IStackable {
	return s.Stack
}

func (s *Server) UseRouter(path string, router *Router) {
	router.Path = path
	s.Stack = append(s.Stack, router)
}

func (s *Server) Listen() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("could not start server: %v", err)
	}
	defer ln.Close()
	s.listener = ln

	fmt.Println("Server is running on", s.addr)

	doneCh := make(chan struct{})
	go s.acceptLoop()

	<-doneCh
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn)
	}
}

func currentDateString() string {
	time := time.Now().UTC().Format(time.RFC1123)
	return "Date: " + time[:len(time)-3] + "GMT"
}

func (s *Server) handleConnection(conn net.Conn) {
	req := NewRequest()

	err := req.parseRequest(conn)
	if err != nil {
		return
	}

	res := &Response{
		Server:  s,
		conn:    conn,
		Body:    []byte{},
		Status:  200,
		Headers: ResponseHeaders{},
	}
	var result []IStackable

	if s.CorsOptions != nil {
		for _, v := range s.CorsOptions.Origins {
			if v == req.Headers["origin"] {
				for k, h := range s.CorsOptions.Headers {
					res.Headers.Add(k, h)
				}
			}
		}
	}

	if req.Method == "OPTIONS" {
		res.Status = 204
		res.Body = nil
		res.Done()
		return
	}

	var found bool
	traverseStackables(req, s, "", &result, &found)
	if !found {
		res.ApiError(404, "Unknown route.")
		res.Done()
		err = errors.New("Not found!")
		return
	}
outer:
	for _, h := range result {
		switch s := h.(type) {
		case Route:
			{
				req.parseBody()
				err := s.Handler(res, req)
				if err != nil {
					res.ApiError(500, err.Error())
					res.Done()
					return
				}
				break outer
			}
		case Middleware:
			{
				err := s.Handler(res, req)
				if err != nil {
					res.ApiError(500, err.Error())
					res.Done()
					return
				}
			}
		default:
			{
				continue
			}
		}
	}

	if res.EventStream != nil {
		s.EventStreams[res.EventStream.Identifier] = res.EventStream.Ch
		res.Headers.Add("X-Accel-Buffering", "no") //ngonanoweb bs
		go res.StreamEvents()
	} else {
		res.Done()
	}
}

func removeEmpty(v string) bool {
	return v == ""
}

func traverseStackables(req *Request, stackable IStackable, parentPath string, result *[]IStackable, found *bool) {
	if *found {
		return
	}
	switch s := stackable.(type) {
	case Route:
		if s.Method != req.Method {
			return
		}

		fullPath := strings.TrimSuffix(parentPath, "/") + "/" + strings.TrimPrefix(s.Path, "/")

		routeParts := slices.DeleteFunc(strings.Split(fullPath, "/"), removeEmpty)
		pathParts := slices.DeleteFunc(strings.Split(req.Path, "/"), removeEmpty)
		if len(routeParts) != len(pathParts) {
			return
		}

		params := make(map[string]string)
		match := true
		for i, routePart := range routeParts {
			pathPart := pathParts[i]

			if strings.HasPrefix(routePart, ":") {
				paramName := routePart[1:]
				params[paramName] = pathPart
			} else {
				if routePart != pathPart {
					match = false
					break
				}
			}
		}

		if match {
			req.Params = params
			*found = true
			*result = append(*result, stackable)
			return
		}

	case *Router:
		{
			parentPath = strings.TrimSuffix(parentPath, "/") + "/" + strings.TrimPrefix(s.Path, "/")

			for _, s := range stackable.GetStack() {
				traverseStackables(req, s, parentPath, result, found)
			}
		}
	case Middleware:
		{
			*result = append(*result, stackable)
		}
	case *Server:
		{
			parentPath = strings.TrimSuffix(parentPath, "/") + "/"
			for _, s := range stackable.GetStack() {
				traverseStackables(req, s, parentPath, result, found)
			}
		}
	}

}

func (s *Server) SendEvent(identifier string, message string) {
	if ch, ok := s.EventStreams[identifier]; ok {
		select {
		case *ch <- message:
		default:
			delete(s.EventStreams, identifier)
			close(*ch)
		}
	}
}

type ApiError struct {
	StatusCode int
	Message    string
}
