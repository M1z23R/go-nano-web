package gonanoweb

type IStackable interface {
	GetStack() []IStackable
}

type Handler func(res *Response, req *Request) error

type IHandler interface {
	Handler
}

type Middleware struct {
	IStackable
	Handler Handler
}

func (m Middleware) GetStack() []IStackable {
	return []IStackable{}
}

type Route struct {
	IStackable
	Path    string
	Method  string
	Handler Handler
}

func (r Route) GetStack() []IStackable {
	return []IStackable{}
}
