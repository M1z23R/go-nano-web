package gonanoweb

func (r *Router) Get(path string, handler Handler, middlewares ...Middleware) {
	for _, m := range middlewares {
		r.Stack = append(r.Stack, m)
	}

	r.Stack = append(r.Stack, Route{Path: path, Handler: handler, Method: "GET"})
}

func (r *Router) Post(path string, handler Handler, middlewares ...Middleware) {
	for _, m := range middlewares {
		r.Stack = append(r.Stack, m)
	}
	r.Stack = append(r.Stack, Route{Path: path, Handler: handler, Method: "POST"})
}

func (r *Router) Put(path string, handler Handler, middlewares ...Middleware) {
	for _, m := range middlewares {
		r.Stack = append(r.Stack, m)
	}

	r.Stack = append(r.Stack, Route{Path: path, Handler: handler, Method: "PUT"})
}

func (r *Router) Patch(path string, handler Handler, middlewares ...Middleware) {
	for _, m := range middlewares {
		r.Stack = append(r.Stack, m)
	}

	r.Stack = append(r.Stack, Route{Path: path, Handler: handler, Method: "PATCH"})
}

func (r *Router) Delete(path string, handler Handler, middlewares ...Middleware) {
	for _, m := range middlewares {
		r.Stack = append(r.Stack, m)
	}

	r.Stack = append(r.Stack, Route{Path: path, Handler: handler, Method: "DELETE"})
}
