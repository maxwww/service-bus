package bus

import (
	"fmt"
)

type HandlerFunc func(*Msg)

type MiddlewareFunc func(HandlerFunc) HandlerFunc

type Handler interface {
	ServeRMQ(*Msg)
}

type Route struct {
	handler HandlerFunc
	path    string
}

type Router struct {
	routes         []*Route
	handler        HandlerFunc
	defaultHandler HandlerFunc
	middlewares    []MiddlewareFunc
}

func NewRouter() *Router {
	router := &Router{}
	router.handler = router.serveRMQ

	return router
}

func (r *Router) ServeRMQ(d *Msg) {
	r.handler(d)
}

func (r *Router) serveRMQ(d *Msg) {
	if path, ok := d.Delivery.Headers["path"]; ok {
		var handler HandlerFunc
		for _, route := range r.routes {
			if path == route.path {
				handler = route.handler
				break
			}
		}
		if handler != nil {
			handler(d)
		} else if r.defaultHandler != nil {
			r.defaultHandler(d)
		} else {
			fmt.Println("404. not found")
		}
	}
}

func (r *Router) Handle(path string, handler HandlerFunc) {
	for _, route := range r.routes {
		if path == route.path {
			route.handler = handler
			return
		}
	}

	r.routes = append(r.routes, &Route{
		handler: handler,
		path:    path,
	})
}

func (r *Router) Use(middleware MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middleware)

	chainHandlers := r.serveRMQ

	for i := range r.middlewares {
		chainHandlers = r.middlewares[len(r.middlewares)-1-i](chainHandlers)
	}

	r.handler = chainHandlers
}

func (r *Router) HandleDefault(handler HandlerFunc) {
	r.defaultHandler = handler
}
