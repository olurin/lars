package lars

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

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

type testGlobs struct {
	Context
}

func (g *testGlobs) Reset() {

}

func TestContextPath(t *testing.T) {
	l := New()
	r := l.router

	r.add(GET, "/users/:id", nil, l)
	c := l.pool.New().(*Context)

	r.Find(GET, "/users/1", c)
	Equal(t, c.Path(), "/users/:id")

	r.add(GET, "/users/:uid/files/:fid", nil, l)
	c = l.pool.New().(*Context)

	r.Find(GET, "/users/1/files/1", c)
	Equal(t, c.Path(), "/users/:uid/files/:fid")
}

func TestGlobals(t *testing.T) {

	l := New()
	l.RegisterGlobalsFunc(func() IGlobals { return &testGlobs{} })

	// Route
	l.Get("/", func(c *Context) {
		c.Response.Write([]byte(fmt.Sprint(reflect.TypeOf(c.Globals))))
	})

	c, b := request(GET, "/", l)
	Equal(t, http.StatusOK, c)
	Equal(t, b, "*lars.testGlobs")
}

func TestGetSet(t *testing.T) {
	l := New()
	r := l.router
	path := "/folders/a/files/echo.gif"
	r.add(GET, path, func(c *Context) {
		c.Set("path", path)
	}, l)

	c := l.pool.New().(*Context)
	h, _ := r.Find(GET, path, c)
	NotEqual(t, h, nil)
	h(c)
	Equal(t, path, c.Get("path"))

	c.store = nil
	c.Set("a", 1)
	Equal(t, 1, c.Get("a"))
}

func TestContextGolangContext(t *testing.T) {
	l := New()
	c := l.pool.New().(*Context)
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"foo\":\"bar\", \"bar\":\"foo\"}"))

	Equal(t, c.Err(), nil)
	Equal(t, c.Done(), nil)

	ti, ok := c.Deadline()
	Equal(t, ti, time.Time{})
	Equal(t, ok, false)
	Equal(t, c.Value(0), c.Request)
	Equal(t, c.Value("foo"), nil)

	c.Set("foo", "bar")
	Equal(t, c.Value("foo"), "bar")
	Equal(t, c.Value(1), nil)
}

func TestContextNetContext(t *testing.T) {
	l := New()
	c := l.pool.New().(*Context)
	c.Set("key", "val")
	Equal(t, "val", c.Value("key"))
}
