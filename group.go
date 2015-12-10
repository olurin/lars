package lars

// IRouteGroup interface for router group
type IRouteGroup interface {
	IRoutes
	Group(prefix string, m ...Middleware) IRouteGroup
}

// IRoutes interface for routes
type IRoutes interface {
	Use(...Middleware)
	Any(string, Handler)
	Get(string, Handler)
	Post(string, Handler)
	Delete(string, Handler)
	Patch(string, Handler)
	Put(string, Handler)
	Options(string, Handler)
	Head(string, Handler)
	Connect(string, Handler)
	Trace(string, Handler)
}

// RouteGroup struct containing all fields and methods for use.
type RouteGroup struct {
	lars *LARS
}

var _ IRouteGroup = &RouteGroup{}

// Use adds a middleware handler to the group middleware chain.
func (g *RouteGroup) Use(m ...Middleware) {
	for _, h := range m {
		g.lars.middleware = append(g.lars.middleware, wrapMiddleware(h))
	}
}

// Connect adds a CONNECT route & handler to the router.
func (g *RouteGroup) Connect(path string, h Handler) {
	g.lars.add(CONNECT, path, h)
}

// Delete adds a DELETE route & handler to the router.
func (g *RouteGroup) Delete(path string, h Handler) {
	g.lars.add(DELETE, path, h)
}

// Get adds a GET route & handler to the router.
func (g *RouteGroup) Get(path string, h Handler) {
	g.lars.add(GET, path, h)
}

// Head adds a HEAD route & handler to the router.
func (g *RouteGroup) Head(path string, h Handler) {
	g.lars.add(HEAD, path, h)
}

// Options adds an OPTIONS route & handler to the router.
func (g *RouteGroup) Options(path string, h Handler) {
	g.lars.add(OPTIONS, path, h)
}

// Patch adds a PATCH route & handler to the router.
func (g *RouteGroup) Patch(path string, h Handler) {
	g.lars.add(PATCH, path, h)
}

// Post adds a POST route & handler to the router.
func (g *RouteGroup) Post(path string, h Handler) {
	g.lars.add(POST, path, h)
}

// Put adds a PUT route & handler to the router.
func (g *RouteGroup) Put(path string, h Handler) {
	g.lars.add(PUT, path, h)
}

// Trace adds a TRACE route & handler to the router.
func (g *RouteGroup) Trace(path string, h Handler) {
	g.lars.add(TRACE, path, h)
}

// Any adds a route & handler to the router for all HTTP methods.
func (g *RouteGroup) Any(path string, h Handler) {
	for _, m := range methods {
		g.lars.add(m, path, h)
	}
}

// Match adds a route & handler to the router for multiple HTTP methods provided.
func (g *RouteGroup) Match(methods []string, path string, h Handler) {
	for _, m := range methods {
		g.lars.add(m, path, h)
	}
}

// Group creates a new sub router with prefix. It inherits all properties from
// the parent. Passing middleware overrides parent middleware.
func (g *RouteGroup) Group(prefix string, m ...Middleware) IRouteGroup {
	l := *g.lars
	ng := &RouteGroup{&l}
	ng.lars.prefix += prefix

	if len(m) == 0 {
		mw := make([]MiddlewareFunc, len(ng.lars.middleware))
		copy(mw, ng.lars.middleware)
		ng.lars.middleware = mw

		return ng
	}

	ng.lars.middleware = nil
	ng.Use(m...)

	return ng
}

// type Group struct {
// 	lars LARS
// }

// func (g *Group) Use(m ...Middleware) {
// 	for _, h := range m {
// 		g.lars.middleware = append(g.lars.middleware, wrapMiddleware(h))
// 	}
// }

// func (g *Group) Connect(path string, h Handler) {
// 	g.lars.Connect(path, h)
// }

// func (g *Group) Delete(path string, h Handler) {
// 	g.lars.Delete(path, h)
// }

// func (g *Group) Get(path string, h Handler) {
// 	g.lars.Get(path, h)
// }

// func (g *Group) Head(path string, h Handler) {
// 	g.lars.Head(path, h)
// }

// func (g *Group) Options(path string, h Handler) {
// 	g.lars.Options(path, h)
// }

// func (g *Group) Patch(path string, h Handler) {
// 	g.lars.Patch(path, h)
// }

// func (g *Group) Post(path string, h Handler) {
// 	g.lars.Post(path, h)
// }

// func (g *Group) Put(path string, h Handler) {
// 	g.lars.Put(path, h)
// }

// func (g *Group) Trace(path string, h Handler) {
// 	g.lars.Trace(path, h)
// }

// // func (g *Group) WebSocket(path string, h HandlerFunc) {
// // 	g.lars.WebSocket(path, h)
// // }

// // func (g *Group) Static(path, root string) {
// // 	g.lars.Static(path, root)
// // }

// // func (g *Group) ServeDir(path, root string) {
// // 	g.lars.ServeDir(path, root)
// // }

// // func (g *Group) ServeFile(path, file string) {
// // 	g.lars.ServeFile(path, file)
// // }

// func (g *Group) Group(prefix string, m ...Middleware) *Group {
// 	return g.lars.Group(prefix, m...)
// }
