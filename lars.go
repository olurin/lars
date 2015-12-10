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

	notFoundHandler = func(c *Context) {
		http.Error(c.Response, "4040 not found", http.StatusNotFound)
	}

	methodNotAllowedHandler = func(c *Context) {
		http.Error(c.Response, "4040 not allowed", http.StatusMethodNotAllowed)
	}
)

// New creates an instance of lars.
func New() *LARS {
	e := &LARS{maxParam: new(int)}
	e.RouteGroup = RouteGroup{e}
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
