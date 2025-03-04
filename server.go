package gonanoweb

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"
)

type ServerOptions struct {
	ReadTimeout     *time.Duration
	WriteTimeout    *time.Duration
	CorsOptions     *CorsOptions
	MaxRequestSize  *int64
	TLSConfig       *tls.Config
	SecurityHeaders *bool
}

type Server struct {
	addr            string
	listener        net.Listener
	Stack           []IStackable
	EventStreams    map[string]*chan string
	CorsOptions     *CorsOptions
	TLSConfig       *tls.Config
	ReadTimeout     *time.Duration
	WriteTimeout    *time.Duration
	MaxRequestSize  *int64
	SecurityHeaders *bool
	FormDataOptions *FormDataOptions
}

func NewServer(addr string, options *ServerOptions) *Server {
	server := &Server{
		EventStreams: make(map[string]*chan string),
		Stack:        []IStackable{},
		addr:         addr,
	}

	if options != nil {
		if options.ReadTimeout != nil {
			server.ReadTimeout = options.ReadTimeout
		}
		if options.WriteTimeout != nil {
			server.WriteTimeout = options.WriteTimeout
		}
		if options.CorsOptions != nil {
			server.CorsOptions = options.CorsOptions
		}
		if options.MaxRequestSize != nil {
			server.MaxRequestSize = options.MaxRequestSize
		}
		if options.TLSConfig != nil {
			server.TLSConfig = options.TLSConfig
		}
		if options.SecurityHeaders != nil {
			server.SecurityHeaders = options.SecurityHeaders
		}
	}

	return server
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

func (s *Server) ListenTLS(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	ln, err := tls.Listen("tcp", s.addr, config)
	if err != nil {
		return err
	}
	s.listener = ln
	return s.Listen()
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

func (s *Server) handleConnection(conn net.Conn) {
	if s.ReadTimeout != nil && *s.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(*s.ReadTimeout))
	}
	if s.WriteTimeout != nil && *s.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(*s.WriteTimeout))
	}

	req := NewRequest()
	req.server = s
	req.MaxRequestSize = s.MaxRequestSize
	req.conn = &conn
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

	s.handleCORS(res, req)

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
		res.Headers.Add("X-Accel-Buffering", "no") //nginx bs
		go res.StreamEvents()
	} else {
		res.Done()
	}
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
