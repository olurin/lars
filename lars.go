package lars

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sync"
)

// LARS struct containing all fields and methods for use
type LARS struct {
	RouteGroup
	prefix        string
	middleware    []MiddlewareFunc
	maxParam      *int
	pool          sync.Pool
	router        *router
	http404       HandlerFunc
	newGlobals    GlobalsFunc
	globalsExists bool

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	// Order of checks:
	// > Attempts to find the same path but all lowercase
	// > Attempts to find by adding or removing slash
	// > Falls Back to Not Found Handler
	FixTrailingSlash bool
}

type route struct {
	Method  string
	Path    string
	Handler Handler
}

// Middleware is the type used in registerig middleware.
// NOTE: these middlewares may get wrapped by the MiddlewareFunc
// type for chainging purposes internally.
type Middleware interface{}

// MiddlewareFunc is the final type used for the middleware and chaining of it
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Handler is the type used in registering handlers.
// NOTE: these handlers may get wrapped by the HandlerFunc
// type internally.
type Handler interface{}

// HandlerFunc is the internal handler type used for handlers.
type HandlerFunc func(*Context)

// GlobalsFunc is a function that creates a new Global object to be passed around the request
type GlobalsFunc func() IGlobals

// HTTP Constant Terms and Variables
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

	default404Body = "404 page not found"
	default405Body = "405 method not allowed"

	basePath = "/"
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

	//----------------
	// Error handlers
	//----------------

	defaultNotFoundHandler = func(c *Context) {
		http.Error(c.Response, default404Body, http.StatusNotFound)
	}

	methodNotAllowedHandler = func(c *Context) {
		http.Error(c.Response, default405Body, http.StatusMethodNotAllowed)
	}
)

// New creates an instance of lars.
// FixTrailingSlash defaults to true
func New() *LARS {
	l := &LARS{
		FixTrailingSlash: true,
		maxParam:         new(int),
		http404:          defaultNotFoundHandler,
		newGlobals: func() IGlobals {
			return nil
		},
	}
	l.RouteGroup = RouteGroup{l}
	l.pool.New = func() interface{} {
		return &Context{
			Request:       nil,
			Response:      new(Response),
			pvalues:       make([]string, *l.maxParam),
			store:         make(store),
			Globals:       l.newGlobals(),
			globalsExists: l.globalsExists,
		}
	}
	l.router = newRouter(l)

	return l
}

// RegisterNotFoundFunc alows for overriding of the not found handler function.
// Here can set redirecting to about to about/ or about/ to about
// and all your other SEO needs
func (l *LARS) RegisterNotFoundFunc(notFound HandlerFunc) {
	l.http404 = notFound
}

// RegisterGlobalsFunc registers a custom globals function for creation
// and resetting of a global object passed per http request
func (l *LARS) RegisterGlobalsFunc(fn GlobalsFunc) {
	l.newGlobals = fn
	l.globalsExists = true
}

func (l *LARS) add(method, path string, h Handler) {
	path = l.prefix + path
	l.router.add(method, path, wrapHandler(h), l)
	r := route{
		Method:  method,
		Path:    path,
		Handler: runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name(),
	}
	l.router.routes = append(l.router.routes, r)
}

// URI generates a URI from handler.
func (l *LARS) URI(h Handler, params ...interface{}) string {
	uri := new(bytes.Buffer)
	pl := len(params)
	n := 0
	hn := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	for _, r := range l.router.routes {
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
func (l *LARS) URL(h Handler, params ...interface{}) string {
	return l.URI(h, params...)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (l *LARS) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	c := l.pool.Get().(*Context)
	h, l := l.router.Find(r.Method, r.URL.Path, c)
	c.reset(r, w, l)

	// Chain middleware with handler in the end
	for i := len(l.middleware) - 1; i >= 0; i-- {
		h = l.middleware[i](h)
	}

	// Execute chain
	h(c)

	l.pool.Put(c)
}
