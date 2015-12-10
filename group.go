package lars

type Group struct {
	lars LARS
}

func (g *Group) Use(m ...Middleware) {
	for _, h := range m {
		g.lars.middleware = append(g.lars.middleware, wrapMiddleware(h))
	}
}

func (g *Group) Connect(path string, h Handler) {
	g.lars.Connect(path, h)
}

func (g *Group) Delete(path string, h Handler) {
	g.lars.Delete(path, h)
}

func (g *Group) Get(path string, h Handler) {
	g.lars.Get(path, h)
}

func (g *Group) Head(path string, h Handler) {
	g.lars.Head(path, h)
}

func (g *Group) Options(path string, h Handler) {
	g.lars.Options(path, h)
}

func (g *Group) Patch(path string, h Handler) {
	g.lars.Patch(path, h)
}

func (g *Group) Post(path string, h Handler) {
	g.lars.Post(path, h)
}

func (g *Group) Put(path string, h Handler) {
	g.lars.Put(path, h)
}

func (g *Group) Trace(path string, h Handler) {
	g.lars.Trace(path, h)
}

// func (g *Group) WebSocket(path string, h HandlerFunc) {
// 	g.lars.WebSocket(path, h)
// }

// func (g *Group) Static(path, root string) {
// 	g.lars.Static(path, root)
// }

// func (g *Group) ServeDir(path, root string) {
// 	g.lars.ServeDir(path, root)
// }

// func (g *Group) ServeFile(path, file string) {
// 	g.lars.ServeFile(path, file)
// }

func (g *Group) Group(prefix string, m ...Middleware) *Group {
	return g.lars.Group(prefix, m...)
}
