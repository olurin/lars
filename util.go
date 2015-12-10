package lars

import "net/http"

// wrapMiddleware wraps middleware.
func wrapMiddleware(m Middleware) MiddlewareFunc {
	switch m := m.(type) {
	case MiddlewareFunc:
		return m
	case func(HandlerFunc) HandlerFunc:
		return m
	case HandlerFunc:
		return wrapHandlerFuncMW(m)
	case func(*Context):
		return wrapHandlerFuncMW(m)
	case func(http.Handler) http.Handler:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) {
				m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					h(c)
				})).ServeHTTP(c.Response, c.Request)
			}
		}
	case http.Handler:
		return wrapHTTPHandlerFuncMW(m.ServeHTTP)
	case func(http.ResponseWriter, *http.Request):
		return wrapHTTPHandlerFuncMW(m)
	default:
		panic("unknown middleware")
	}
}

// wrapHandlerFuncMW wraps HandlerFunc middleware.
func wrapHandlerFuncMW(m HandlerFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) {
			if m(c); c.Response.status != http.StatusOK || c.Response.committed {
				return
			}
			h(c)
		}
	}
}

// wrapHTTPHandlerFuncMW wraps http.HandlerFunc middleware.
func wrapHTTPHandlerFuncMW(m http.HandlerFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) {
			if !c.Response.committed {
				m.ServeHTTP(c.Response, c.Request)
			}
			h(c)
		}
	}
}

// wrapHandler wraps handler.
func wrapHandler(h Handler) HandlerFunc {
	switch h := h.(type) {
	case HandlerFunc:
		return h
	case func(*Context):
		return h
	case http.Handler, http.HandlerFunc:
		return func(c *Context) {
			h.(http.Handler).ServeHTTP(c.Response, c.Request)
		}
	case func(http.ResponseWriter, *http.Request):
		return func(c *Context) {
			h(c.Response, c.Request)
		}
	default:
		panic("unknown handler")
	}
}
