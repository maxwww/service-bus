package bus

type HandlerFunc func(*Msg)

type Handler interface {
	ServeRMQ(*Msg)
}

type Route struct {
	handler HandlerFunc
	path    string
}
type Router struct {
	routes []*Route
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) ServeRMQ(d *Msg) {
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
