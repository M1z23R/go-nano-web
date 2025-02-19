package gonanoweb

func (s *Server) Get(path string, handler Handler, middlewares ...Middleware) {
	validatePath(path)
	for _, m := range middlewares {
		s.Stack = append(s.Stack, m)
	}

	s.Stack = append(s.Stack, Route{Path: path, Handler: handler, Method: "GET"})
}

func (s *Server) Post(path string, handler Handler, middlewares ...Middleware) {
	validatePath(path)
	for _, m := range middlewares {
		s.Stack = append(s.Stack, m)
	}

	s.Stack = append(s.Stack, Route{Path: path, Handler: handler, Method: "POST"})
}

func (s *Server) Put(path string, handler Handler, middlewares ...Middleware) {
	validatePath(path)
	for _, m := range middlewares {
		s.Stack = append(s.Stack, m)
	}

	s.Stack = append(s.Stack, Route{Path: path, Handler: handler, Method: "PUT"})
}

func (s *Server) Patch(path string, handler Handler, middlewares ...Middleware) {
	validatePath(path)
	for _, m := range middlewares {
		s.Stack = append(s.Stack, m)
	}

	s.Stack = append(s.Stack, Route{Path: path, Handler: handler, Method: "PATCH"})
}

func (s *Server) Delete(path string, handler Handler, middlewares ...Middleware) {
	validatePath(path)
	for _, m := range middlewares {
		s.Stack = append(s.Stack, m)
	}

	s.Stack = append(s.Stack, Route{Path: path, Handler: handler, Method: "DELETE"})
}
