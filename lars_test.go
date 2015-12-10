package lars

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	. "gopkg.in/go-playground/assert.v1"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

func TestMain(m *testing.M) {

	// setup

	os.Exit(m.Run())

	// teardown
}

func Testl(t *testing.T) {
	l := New()
	NotEqual(t, l, nil)
	NotEqual(t, l.router, nil)
	NotEqual(t, l.pool, nil)
	Equal(t, l.maxParam, 0)

	i := 25
	l.maxParam = &i
	Equal(t, l.maxParam, 25)
}

func TestParamIncrementing(t *testing.T) {
	l := New()
	i := 0
	l.maxParam = &i
	r := l.router

	r.add(GET, "/users/:id", func(*Context) {}, l)

	Equal(t, l.maxParam, 1)
}

func TestMiddleware(t *testing.T) {
	l := New()
	buf := new(bytes.Buffer)

	// trafficcop.MiddlewareFunc
	l.Use(MiddlewareFunc(func(h HandlerFunc) HandlerFunc {
		return func(c *Context) {
			buf.WriteString("a")
			h(c)
		}
	}))

	// func(trafficcop.HandlerFunc) trafficcop.HandlerFunc
	l.Use(func(h HandlerFunc) HandlerFunc {
		return func(c *Context) {
			buf.WriteString("b")
			h(c)
		}
	})

	// trafficcop.HandlerFunc
	l.Use(HandlerFunc(func(c *Context) {
		buf.WriteString("c")
	}))

	// func(trafficcop.Context) error
	l.Use(func(c *Context) {
		buf.WriteString("d")
	})

	// func(http.Handler) http.Handler
	l.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf.WriteString("e")
			h.ServeHTTP(w, r)
		})
	})

	// http.Handler
	l.Use(http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteString("f")
	})))

	// http.HandlerFunc
	l.Use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteString("g")
	}))

	// func(http.ResponseWriter, *http.Request)
	l.Use(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteString("h")
	})

	PanicMatches(t, func() { l.Use(nil) }, "unknown middleware")

	// Route
	l.Get("/", func(c *Context) {
		c.Response.Write([]byte("Hello!"))
	})

	c, b := request(GET, "/", l)
	Equal(t, "abcdefgh", buf.String())
	Equal(t, http.StatusOK, c)
	Equal(t, "Hello!", b)

	// Error
	l.Use(func(c *Context) {
		http.Error(c.Response, "500 Internal Server Error", http.StatusInternalServerError)
	})

	c, b = request(GET, "/", l)
	Equal(t, http.StatusInternalServerError, c)
}

func TestHandler(t *testing.T) {
	l := New()

	// HandlerFunc
	l.Get("/1", HandlerFunc(func(c *Context) {
		c.Response.Write([]byte("1"))
	}))

	// func(trafficcop.Context) error
	l.Get("/2", func(c *Context) {
		c.Response.Write([]byte("2"))
	})

	// http.Handler/http.HandlerFunc
	l.Get("/3", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("3"))
	}))

	// func(http.ResponseWriter, *http.Request)
	l.Get("/4", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("4"))
	})

	for _, p := range []string{"1", "2", "3", "4"} {
		c, b := request(GET, "/"+p, l)
		Equal(t, http.StatusOK, c)
		Equal(t, p, b)
	}

	// Unknown
	PanicMatches(t, func() { l.Get("/5", nil) }, "unknown handler")
}

func TestConnect(t *testing.T) {
	l := New()
	testMethod(t, CONNECT, "/", l)
}

func TestDelete(t *testing.T) {
	l := New()
	testMethod(t, DELETE, "/", l)
}

func TestGet(t *testing.T) {
	l := New()
	testMethod(t, GET, "/", l)
}

func TestHead(t *testing.T) {
	l := New()
	testMethod(t, HEAD, "/", l)
}

func TestOptions(t *testing.T) {
	l := New()
	testMethod(t, OPTIONS, "/", l)
}

func TestPatch(t *testing.T) {
	l := New()
	testMethod(t, PATCH, "/", l)
}

func TestPost(t *testing.T) {
	l := New()
	testMethod(t, POST, "/", l)
}

func TestPut(t *testing.T) {
	l := New()
	testMethod(t, PUT, "/", l)
}

func TestTrace(t *testing.T) {
	l := New()
	testMethod(t, TRACE, "/", l)
}

func TestAny(t *testing.T) {
	l := New()
	l.Any("/", func(c *Context) {
		c.Response.Write([]byte("Any"))
	})
}

func TestMatch(t *testing.T) {
	l := New()
	l.Match([]string{GET, POST}, "/", func(c *Context) {
		c.Response.Write([]byte("Match"))
	})
}

func TestURL(t *testing.T) {
	l := New()

	static := func(*Context) {}
	getUser := func(*Context) {}
	getFile := func(*Context) {}

	l.Get("/static/file", static)
	l.Get("/users/:id", getUser)
	g := l.Group("/group")
	g.Get("/users/:uid/files/:fid", getFile)

	Equal(t, "/static/file", l.URI(static))
	Equal(t, "/users/:id", l.URL(getUser))
	Equal(t, "/users/1", l.URL(getUser, "1"))
	Equal(t, "/group/users/1/files/:fid", l.URI(getFile, "1"))
	Equal(t, "/group/users/1/files/1", l.URI(getFile, "1", "1"))
}

func TestRoutes(t *testing.T) {
	l := New()
	h := func(*Context) {}
	routes := []route{
		{GET, "/users/:user/events", h},
		{GET, "/users/:user/events/public", h},
		{POST, "/repos/:owner/:repo/git/refs", h},
		{POST, "/repos/:owner/:repo/git/tags", h},
	}

	for _, r := range routes {
		l.add(r.Method, r.Path, h)
	}

	for _, r := range routes {
		c, s := request(r.Method, r.Path, l)
		Equal(t, c, 200)
		Equal(t, s, "")
	}
}

func TestlGroup(t *testing.T) {
	l := New()
	buf := new(bytes.Buffer)

	l.Use(func(*Context) {
		buf.WriteString("0")
	})
	h := func(*Context) {}

	//--------
	// Routes
	//--------

	l.Get("/users", h)

	// Group
	g1 := l.Group("/group1")
	g1.Use(func(*Context) {
		buf.WriteString("1")

	})
	g1.Get("/", h)

	// Group with no parent middleware
	g2 := l.Group("/group2", func(*Context) {
		buf.WriteString("2")
	})
	g2.Get("/", h)

	// Nested groups
	g3 := l.Group("/group3")
	g4 := g3.Group("/group4")
	g4.Get("/", func(c *Context) {})

	request(GET, "/users", l)
	Equal(t, "0", buf.String())

	buf.Reset()
	request(GET, "/group1/", l)
	Equal(t, "01", buf.String())

	buf.Reset()
	request(GET, "/group2/", l)
	Equal(t, "2", buf.String())

	buf.Reset()
	c, _ := request(GET, "/group3/group4/", l)
	Equal(t, http.StatusOK, c)
}

func TestNotFound(t *testing.T) {
	l := New()
	r, _ := http.NewRequest(GET, "/files", nil)
	w := httptest.NewRecorder()

	l.ServeHTTP(w, r)
	Equal(t, http.StatusNotFound, w.Code)
	Equal(t, w.Body.String(), "404 page not found\n")

	// custom not found
	l.RegisterNotFoundFunc(func(ctx *Context) {
		http.Error(ctx.Response, "Custom Not Found", http.StatusNotFound)
	})

	w = httptest.NewRecorder()
	l.ServeHTTP(w, r)
	Equal(t, http.StatusNotFound, w.Code)
	Equal(t, w.Body.String(), "Custom Not Found\n")
}

func TestMethodNotAllowed(t *testing.T) {
	l := New()
	l.Get("/", func(ctx *Context) {
		ctx.Response.Write([]byte("OK!"))
	})

	r, _ := http.NewRequest(POST, "/", nil)
	w := httptest.NewRecorder()

	l.ServeHTTP(w, r)
	Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func testMethod(t *testing.T, method, path string, l *LARS) {

	m := fmt.Sprintf("%c%s", method[0], strings.ToLower(method[1:]))
	p := reflect.ValueOf(path)
	h := reflect.ValueOf(func(c *Context) {
		c.Response.Write([]byte(method))
	})
	i := interface{}(l)
	reflect.ValueOf(i).MethodByName(m).Call([]reflect.Value{p, h})
	_, body := request(method, path, l)
	Equal(t, body, method)
}

func request(method, path string, l *LARS) (int, string) {

	r, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	l.ServeHTTP(w, r)

	return w.Code, w.Body.String()
}
