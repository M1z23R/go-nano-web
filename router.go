package gonanoweb

type Router struct {
	Path  string
	Stack []IStackable
	IStackable
}

func NewRouter() *Router {
	return &Router{
		Stack: []IStackable{},
	}
}

func (r Router) GetStack() []IStackable {
	return r.Stack
}

func (r *Router) UseRouter(path string, router *Router) {
	router.Path = path
	r.Stack = append(r.Stack, router)
}

func (r *Router) UseMiddleware(middleware Middleware) {
	r.Stack = append(r.Stack, middleware)
}

func (r *Server) UseMiddleware(middleware Middleware) {
	r.Stack = append(r.Stack, middleware)
}
