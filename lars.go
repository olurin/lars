package lars

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"
)

type LARS struct {
	prefix     string
	middleware []MiddlewareFunc
	maxParam   *int
	pool       sync.Pool
	router     *router
}

type route struct {
	Method  string
	Path    string
	Handler Handler
}

type Middleware interface{}
type MiddlewareFunc func(HandlerFunc) HandlerFunc
type Handler interface{}
type HandlerFunc func(*Context)

const (
	// CONNECT HTTP method
	CONNECT = "CONNECT"
	// DELETE HTTP method
	DELETE = "DELETE"
	// GET HTTP method
	GET = "GET"
	// HEAD HTTP method
	HEAD = "HEAD"
	// OPTIONS HTTP method
	OPTIONS = "OPTIONS"
	// PATCH HTTP method
	PATCH = "PATCH"
	// POST HTTP method
	POST = "POST"
	// PUT HTTP method
	PUT = "PUT"
	// TRACE HTTP method
	TRACE = "TRACE"

	//-------------
	// Media types
	//-------------

	ApplicationJSON                  = "application/json"
	ApplicationJSONCharsetUTF8       = ApplicationJSON + "; " + CharsetUTF8
	ApplicationJavaScript            = "application/javascript"
	ApplicationJavaScriptCharsetUTF8 = ApplicationJavaScript + "; " + CharsetUTF8
	ApplicationXML                   = "application/xml"
	ApplicationXMLCharsetUTF8        = ApplicationXML + "; " + CharsetUTF8
	ApplicationForm                  = "application/x-www-form-urlencoded"
	ApplicationProtobuf              = "application/protobuf"
	ApplicationMsgpack               = "application/msgpack"
	TextHTML                         = "text/html"
	TextHTMLCharsetUTF8              = TextHTML + "; " + CharsetUTF8
	TextPlain                        = "text/plain"
	TextPlainCharsetUTF8             = TextPlain + "; " + CharsetUTF8
	MultipartForm                    = "multipart/form-data"

	//---------
	// Charset
	//---------

	CharsetUTF8 = "charset=utf-8"

	//---------
	// Headers
	//---------

	AcceptEncoding     = "Accept-Encoding"
	Authorization      = "Authorization"
	ContentDisposition = "Content-Disposition"
	ContentEncoding    = "Content-Encoding"
	ContentLength      = "Content-Length"
	ContentType        = "Content-Type"
	Location           = "Location"
	Upgrade            = "Upgrade"
	Vary               = "Vary"
	WWWAuthenticate    = "WWW-Authenticate"
	XForwardedFor      = "X-Forwarded-For"
	XRealIP            = "X-Real-IP"
	//-----------
	// Protocols
	//-----------

	WebSocket = "websocket"

	indexPage = "index.html"
)

var (
	methods = [...]string{
		CONNECT,
		DELETE,
		GET,
		HEAD,
		OPTIONS,
		PATCH,
		POST,
		PUT,
		TRACE,
	}

	//--------
	// Errors
	//--------

	UnsupportedMediaType  = errors.New("unsupported media type")
	RendererNotRegistered = errors.New("renderer not registered")
	InvalidRedirectCode   = errors.New("invalid redirect status code")

	//----------------
	// Error handlers
	//----------------

	notFoundHandler = func(c *Context) {
		http.Error(c.Response, "4040 not found", http.StatusNotFound)
		// return nil
	}

	methodNotAllowedHandler = func(c *Context) {
		http.Error(c.Response, "4040 not allowed", http.StatusMethodNotAllowed)
		// return nil
		// return NewHTTPError(http.StatusMethodNotAllowed)
	}

	unixEpochTime = time.Unix(0, 0)
)

// New creates an instance of lars.
func New() *LARS {
	e := &LARS{maxParam: new(int)}
	e.pool.New = func() interface{} {
		return &Context{
			Request:  nil,
			Response: new(Response),
			lars:     e,
			pvalues:  make([]string, *e.maxParam),
			store:    make(store),
		}
	}
	e.router = newRouter(e)

	return e
}

// // Router returns router.
// func (e *LARS) Router() *Router {
// 	return e.router
// }

// Use adds handler to the middleware chain.
func (e *LARS) Use(m ...Middleware) {
	for _, h := range m {
		e.middleware = append(e.middleware, wrapMiddleware(h))
	}
}

// Connect adds a CONNECT route > handler to the router.
func (e *LARS) Connect(path string, h Handler) {
	e.add(CONNECT, path, h)
}

// Delete adds a DELETE route > handler to the router.
func (e *LARS) Delete(path string, h Handler) {
	e.add(DELETE, path, h)
}

// Get adds a GET route > handler to the router.
func (e *LARS) Get(path string, h Handler) {
	e.add(GET, path, h)
}

// Head adds a HEAD route > handler to the router.
func (e *LARS) Head(path string, h Handler) {
	e.add(HEAD, path, h)
}

// Options adds an OPTIONS route > handler to the router.
func (e *LARS) Options(path string, h Handler) {
	e.add(OPTIONS, path, h)
}

// Patch adds a PATCH route > handler to the router.
func (e *LARS) Patch(path string, h Handler) {
	e.add(PATCH, path, h)
}

// Post adds a POST route > handler to the router.
func (e *LARS) Post(path string, h Handler) {
	e.add(POST, path, h)
}

// Put adds a PUT route > handler to the router.
func (e *LARS) Put(path string, h Handler) {
	e.add(PUT, path, h)
}

// Trace adds a TRACE route > handler to the router.
func (e *LARS) Trace(path string, h Handler) {
	e.add(TRACE, path, h)
}

// Any adds a route > handler to the router for all HTTP methods.
func (e *LARS) Any(path string, h Handler) {
	for _, m := range methods {
		e.add(m, path, h)
	}
}

// Match adds a route > handler to the router for multiple HTTP methods provided.
func (e *LARS) Match(methods []string, path string, h Handler) {
	for _, m := range methods {
		e.add(m, path, h)
	}
}

func (e *LARS) add(method, path string, h Handler) {
	path = e.prefix + path
	e.router.Add(method, path, wrapHandler(h), e)
	r := route{
		Method:  method,
		Path:    path,
		Handler: runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name(),
	}
	e.router.routes = append(e.router.routes, r)
}

// Group creates a new sub router with prefix. It inherits all properties from
// the parent. Passing middleware overrides parent middleware.
func (e *LARS) Group(prefix string, m ...Middleware) *Group {
	g := &Group{*e}
	g.lars.prefix += prefix
	if len(m) == 0 {
		mw := make([]MiddlewareFunc, len(g.lars.middleware))
		copy(mw, g.lars.middleware)
		g.lars.middleware = mw
	} else {
		g.lars.middleware = nil
		g.Use(m...)
	}
	return g
}

// URI generates a URI from handler.
func (e *LARS) URI(h Handler, params ...interface{}) string {
	uri := new(bytes.Buffer)
	pl := len(params)
	n := 0
	hn := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	for _, r := range e.router.routes {
		if r.Handler == hn {
			for i, l := 0, len(r.Path); i < l; i++ {
				if r.Path[i] == ':' && n < pl {
					for ; i < l && r.Path[i] != '/'; i++ {
					}
					uri.WriteString(fmt.Sprintf("%v", params[n]))
					n++
				}
				if i < l {
					uri.WriteByte(r.Path[i])
				}
			}
			break
		}
	}
	return uri.String()
}

// URL is an alias for `URI` function.
func (e *LARS) URL(h Handler, params ...interface{}) string {
	return e.URI(h, params...)
}

// // Routes returns the registered routes.
// func (e *LARS) Routes() []Route {
// 	return e.router.routes
// }

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *LARS) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	c := e.pool.Get().(*Context)
	h, e := e.router.Find(r.Method, r.URL.Path, c)
	c.reset(r, w, e)

	// Chain middleware with handler in the end
	for i := len(e.middleware) - 1; i >= 0; i-- {
		h = e.middleware[i](h)
	}

	// Execute chain
	h(c)

	e.pool.Put(c)
}

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
