package lars

import (
	"net/http"
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

func TestGroup(t *testing.T) {

	l := New()
	g := l.Group("/group")
	h := func(*Context) {}

	g.Connect("/", h)
	g.Delete("/", h)
	g.Get("/", h)
	g.Head("/", h)
	g.Options("/", h)
	g.Patch("/", h)
	g.Post("/", h)
	g.Put("/", h)
	g.Trace("/", h)

	c, s := request(CONNECT, "/group/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")

	c, s = request(DELETE, "/group/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")

	c, s = request(GET, "/group/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")

	c, s = request(HEAD, "/group/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")

	c, s = request(OPTIONS, "/group/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")

	c, s = request(PATCH, "/group/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")

	c, s = request(POST, "/group/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")

	c, s = request(PUT, "/group/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")

	c, s = request(TRACE, "/group/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")

	c, s = request("BADTYPE", "/group/", l)
	Equal(t, c, http.StatusMethodNotAllowed)
	Equal(t, s, "405 method not allowed\n")

	fn := func() MiddlewareFunc {
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) {
				h(c)
			}
		}
	}

	g2 := l.Group("/othergroup", fn())
	g2.Get("/", h)

	c, s = request(GET, "/othergroup/", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "")
}
