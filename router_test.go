package lars

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

var (
	api = []route{
		// OAuth Authorizations
		{"GET", "/authorizations", nil},
		{"GET", "/authorizations/:id", nil},
		{"POST", "/authorizations", nil},
		//{"PUT", "/authorizations/clients/:client_id", nil},
		//{"PATCH", "/authorizations/:id", nil},
		{"DELETE", "/authorizations/:id", nil},
		{"GET", "/applications/:client_id/tokens/:access_token", nil},
		{"DELETE", "/applications/:client_id/tokens", nil},
		{"DELETE", "/applications/:client_id/tokens/:access_token", nil},

		// Activity
		{"GET", "/events", nil},
		{"GET", "/repos/:owner/:repo/events", nil},
		{"GET", "/networks/:owner/:repo/events", nil},
		{"GET", "/orgs/:org/events", nil},
		{"GET", "/users/:user/received_events", nil},
		{"GET", "/users/:user/received_events/public", nil},
		{"GET", "/users/:user/events", nil},
		{"GET", "/users/:user/events/public", nil},
		{"GET", "/users/:user/events/orgs/:org", nil},
		{"GET", "/feeds", nil},
		{"GET", "/notifications", nil},
		{"GET", "/repos/:owner/:repo/notifications", nil},
		{"PUT", "/notifications", nil},
		{"PUT", "/repos/:owner/:repo/notifications", nil},
		{"GET", "/notifications/threads/:id", nil},
		//{"PATCH", "/notifications/threads/:id", nil},
		{"GET", "/notifications/threads/:id/subscription", nil},
		{"PUT", "/notifications/threads/:id/subscription", nil},
		{"DELETE", "/notifications/threads/:id/subscription", nil},
		{"GET", "/repos/:owner/:repo/stargazers", nil},
		{"GET", "/users/:user/starred", nil},
		{"GET", "/user/starred", nil},
		{"GET", "/user/starred/:owner/:repo", nil},
		{"PUT", "/user/starred/:owner/:repo", nil},
		{"DELETE", "/user/starred/:owner/:repo", nil},
		{"GET", "/repos/:owner/:repo/subscribers", nil},
		{"GET", "/users/:user/subscriptions", nil},
		{"GET", "/user/subscriptions", nil},
		{"GET", "/repos/:owner/:repo/subscription", nil},
		{"PUT", "/repos/:owner/:repo/subscription", nil},
		{"DELETE", "/repos/:owner/:repo/subscription", nil},
		{"GET", "/user/subscriptions/:owner/:repo", nil},
		{"PUT", "/user/subscriptions/:owner/:repo", nil},
		{"DELETE", "/user/subscriptions/:owner/:repo", nil},

		// Gists
		{"GET", "/users/:user/gists", nil},
		{"GET", "/gists", nil},
		//{"GET", "/gists/public", nil},
		//{"GET", "/gists/starred", nil},
		{"GET", "/gists/:id", nil},
		{"POST", "/gists", nil},
		//{"PATCH", "/gists/:id", nil},
		{"PUT", "/gists/:id/star", nil},
		{"DELETE", "/gists/:id/star", nil},
		{"GET", "/gists/:id/star", nil},
		{"POST", "/gists/:id/forks", nil},
		{"DELETE", "/gists/:id", nil},

		// Git Data
		{"GET", "/repos/:owner/:repo/git/blobs/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/blobs", nil},
		{"GET", "/repos/:owner/:repo/git/commits/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/commits", nil},
		//{"GET", "/repos/:owner/:repo/git/refs/*ref", nil},
		{"GET", "/repos/:owner/:repo/git/refs", nil},
		{"POST", "/repos/:owner/:repo/git/refs", nil},
		//{"PATCH", "/repos/:owner/:repo/git/refs/*ref", nil},
		//{"DELETE", "/repos/:owner/:repo/git/refs/*ref", nil},
		{"GET", "/repos/:owner/:repo/git/tags/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/tags", nil},
		{"GET", "/repos/:owner/:repo/git/trees/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/trees", nil},

		// Issues
		{"GET", "/issues", nil},
		{"GET", "/user/issues", nil},
		{"GET", "/orgs/:org/issues", nil},
		{"GET", "/repos/:owner/:repo/issues", nil},
		{"GET", "/repos/:owner/:repo/issues/:number", nil},
		{"POST", "/repos/:owner/:repo/issues", nil},
		//{"PATCH", "/repos/:owner/:repo/issues/:number", nil},
		{"GET", "/repos/:owner/:repo/assignees", nil},
		{"GET", "/repos/:owner/:repo/assignees/:assignee", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/comments", nil},
		//{"GET", "/repos/:owner/:repo/issues/comments", nil},
		//{"GET", "/repos/:owner/:repo/issues/comments/:id", nil},
		{"POST", "/repos/:owner/:repo/issues/:number/comments", nil},
		//{"PATCH", "/repos/:owner/:repo/issues/comments/:id", nil},
		//{"DELETE", "/repos/:owner/:repo/issues/comments/:id", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/events", nil},
		//{"GET", "/repos/:owner/:repo/issues/events", nil},
		//{"GET", "/repos/:owner/:repo/issues/events/:id", nil},
		{"GET", "/repos/:owner/:repo/labels", nil},
		{"GET", "/repos/:owner/:repo/labels/:name", nil},
		{"POST", "/repos/:owner/:repo/labels", nil},
		//{"PATCH", "/repos/:owner/:repo/labels/:name", nil},
		{"DELETE", "/repos/:owner/:repo/labels/:name", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"POST", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name", nil},
		{"PUT", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"GET", "/repos/:owner/:repo/milestones/:number/labels", nil},
		{"GET", "/repos/:owner/:repo/milestones", nil},
		{"GET", "/repos/:owner/:repo/milestones/:number", nil},
		{"POST", "/repos/:owner/:repo/milestones", nil},
		//{"PATCH", "/repos/:owner/:repo/milestones/:number", nil},
		{"DELETE", "/repos/:owner/:repo/milestones/:number", nil},

		// Miscellaneous
		{"GET", "/emojis", nil},
		{"GET", "/gitignore/templates", nil},
		{"GET", "/gitignore/templates/:name", nil},
		{"POST", "/markdown", nil},
		{"POST", "/markdown/raw", nil},
		{"GET", "/meta", nil},
		{"GET", "/rate_limit", nil},

		// Organizations
		{"GET", "/users/:user/orgs", nil},
		{"GET", "/user/orgs", nil},
		{"GET", "/orgs/:org", nil},
		//{"PATCH", "/orgs/:org", nil},
		{"GET", "/orgs/:org/members", nil},
		{"GET", "/orgs/:org/members/:user", nil},
		{"DELETE", "/orgs/:org/members/:user", nil},
		{"GET", "/orgs/:org/public_members", nil},
		{"GET", "/orgs/:org/public_members/:user", nil},
		{"PUT", "/orgs/:org/public_members/:user", nil},
		{"DELETE", "/orgs/:org/public_members/:user", nil},
		{"GET", "/orgs/:org/teams", nil},
		{"GET", "/teams/:id", nil},
		{"POST", "/orgs/:org/teams", nil},
		//{"PATCH", "/teams/:id", nil},
		{"DELETE", "/teams/:id", nil},
		{"GET", "/teams/:id/members", nil},
		{"GET", "/teams/:id/members/:user", nil},
		{"PUT", "/teams/:id/members/:user", nil},
		{"DELETE", "/teams/:id/members/:user", nil},
		{"GET", "/teams/:id/repos", nil},
		{"GET", "/teams/:id/repos/:owner/:repo", nil},
		{"PUT", "/teams/:id/repos/:owner/:repo", nil},
		{"DELETE", "/teams/:id/repos/:owner/:repo", nil},
		{"GET", "/user/teams", nil},

		// Pull Requests
		{"GET", "/repos/:owner/:repo/pulls", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number", nil},
		{"POST", "/repos/:owner/:repo/pulls", nil},
		//{"PATCH", "/repos/:owner/:repo/pulls/:number", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/commits", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/files", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/merge", nil},
		{"PUT", "/repos/:owner/:repo/pulls/:number/merge", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/comments", nil},
		//{"GET", "/repos/:owner/:repo/pulls/comments", nil},
		//{"GET", "/repos/:owner/:repo/pulls/comments/:number", nil},
		{"PUT", "/repos/:owner/:repo/pulls/:number/comments", nil},
		//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number", nil},
		//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number", nil},

		// Repositories
		{"GET", "/user/repos", nil},
		{"GET", "/users/:user/repos", nil},
		{"GET", "/orgs/:org/repos", nil},
		{"GET", "/repositories", nil},
		{"POST", "/user/repos", nil},
		{"POST", "/orgs/:org/repos", nil},
		{"GET", "/repos/:owner/:repo", nil},
		//{"PATCH", "/repos/:owner/:repo", nil},
		{"GET", "/repos/:owner/:repo/contributors", nil},
		{"GET", "/repos/:owner/:repo/languages", nil},
		{"GET", "/repos/:owner/:repo/teams", nil},
		{"GET", "/repos/:owner/:repo/tags", nil},
		{"GET", "/repos/:owner/:repo/branches", nil},
		{"GET", "/repos/:owner/:repo/branches/:branch", nil},
		{"DELETE", "/repos/:owner/:repo", nil},
		{"GET", "/repos/:owner/:repo/collaborators", nil},
		{"GET", "/repos/:owner/:repo/collaborators/:user", nil},
		{"PUT", "/repos/:owner/:repo/collaborators/:user", nil},
		{"DELETE", "/repos/:owner/:repo/collaborators/:user", nil},
		{"GET", "/repos/:owner/:repo/comments", nil},
		{"GET", "/repos/:owner/:repo/commits/:sha/comments", nil},
		{"POST", "/repos/:owner/:repo/commits/:sha/comments", nil},
		{"GET", "/repos/:owner/:repo/comments/:id", nil},
		//{"PATCH", "/repos/:owner/:repo/comments/:id", nil},
		{"DELETE", "/repos/:owner/:repo/comments/:id", nil},
		{"GET", "/repos/:owner/:repo/commits", nil},
		{"GET", "/repos/:owner/:repo/commits/:sha", nil},
		{"GET", "/repos/:owner/:repo/readme", nil},
		//{"GET", "/repos/:owner/:repo/contents/*path", nil},
		//{"PUT", "/repos/:owner/:repo/contents/*path", nil},
		//{"DELETE", "/repos/:owner/:repo/contents/*path", nil},
		//{"GET", "/repos/:owner/:repo/:archive_format/:ref", nil},
		{"GET", "/repos/:owner/:repo/keys", nil},
		{"GET", "/repos/:owner/:repo/keys/:id", nil},
		{"POST", "/repos/:owner/:repo/keys", nil},
		//{"PATCH", "/repos/:owner/:repo/keys/:id", nil},
		{"DELETE", "/repos/:owner/:repo/keys/:id", nil},
		{"GET", "/repos/:owner/:repo/downloads", nil},
		{"GET", "/repos/:owner/:repo/downloads/:id", nil},
		{"DELETE", "/repos/:owner/:repo/downloads/:id", nil},
		{"GET", "/repos/:owner/:repo/forks", nil},
		{"POST", "/repos/:owner/:repo/forks", nil},
		{"GET", "/repos/:owner/:repo/hooks", nil},
		{"GET", "/repos/:owner/:repo/hooks/:id", nil},
		{"POST", "/repos/:owner/:repo/hooks", nil},
		//{"PATCH", "/repos/:owner/:repo/hooks/:id", nil},
		{"POST", "/repos/:owner/:repo/hooks/:id/tests", nil},
		{"DELETE", "/repos/:owner/:repo/hooks/:id", nil},
		{"POST", "/repos/:owner/:repo/merges", nil},
		{"GET", "/repos/:owner/:repo/releases", nil},
		{"GET", "/repos/:owner/:repo/releases/:id", nil},
		{"POST", "/repos/:owner/:repo/releases", nil},
		//{"PATCH", "/repos/:owner/:repo/releases/:id", nil},
		{"DELETE", "/repos/:owner/:repo/releases/:id", nil},
		{"GET", "/repos/:owner/:repo/releases/:id/assets", nil},
		{"GET", "/repos/:owner/:repo/stats/contributors", nil},
		{"GET", "/repos/:owner/:repo/stats/commit_activity", nil},
		{"GET", "/repos/:owner/:repo/stats/code_frequency", nil},
		{"GET", "/repos/:owner/:repo/stats/participation", nil},
		{"GET", "/repos/:owner/:repo/stats/punch_card", nil},
		{"GET", "/repos/:owner/:repo/statuses/:ref", nil},
		{"POST", "/repos/:owner/:repo/statuses/:ref", nil},

		// Search
		{"GET", "/search/repositories", nil},
		{"GET", "/search/code", nil},
		{"GET", "/search/issues", nil},
		{"GET", "/search/users", nil},
		{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword", nil},
		{"GET", "/legacy/repos/search/:keyword", nil},
		{"GET", "/legacy/user/search/:keyword", nil},
		{"GET", "/legacy/user/email/:email", nil},

		// Users
		{"GET", "/users/:user", nil},
		{"GET", "/user", nil},
		//{"PATCH", "/user", nil},
		{"GET", "/users", nil},
		{"GET", "/user/emails", nil},
		{"POST", "/user/emails", nil},
		{"DELETE", "/user/emails", nil},
		{"GET", "/users/:user/followers", nil},
		{"GET", "/user/followers", nil},
		{"GET", "/users/:user/following", nil},
		{"GET", "/user/following", nil},
		{"GET", "/user/following/:user", nil},
		{"GET", "/users/:user/following/:target_user", nil},
		{"PUT", "/user/following/:user", nil},
		{"DELETE", "/user/following/:user", nil},
		{"GET", "/users/:user/keys", nil},
		{"GET", "/user/keys", nil},
		{"GET", "/user/keys/:id", nil},
		{"POST", "/user/keys", nil},
		//{"PATCH", "/user/keys/:id", nil},
		{"DELETE", "/user/keys/:id", nil},
	}
)

func TestRouterStatic(t *testing.T) {
	l := New()
	r := l.router
	path := "/folders/a/files/echo.gif"

	fn := func(c *Context) {
		c.Response.Write([]byte("path:" + path))
	}

	r.add(GET, path, fn, l)

	c := l.pool.New().(*Context)
	h, _ := r.Find(GET, path, c)
	Equal(t, fmt.Sprint(h), fmt.Sprint(fn))
}

func TestRouterParam(t *testing.T) {
	l := New()
	r := l.router
	r.add(GET, "/users/:id", func(c *Context) {}, l)

	c := l.pool.New().(*Context)

	h, _ := r.Find(GET, "/users/1", c)
	NotEqual(t, h, nil)
	Equal(t, "1", c.P(0))
}

func TestRouterTwoParam(t *testing.T) {
	l := New()
	r := l.router
	r.add(GET, "/users/:uid/files/:fid", func(*Context) {}, l)
	c := l.pool.New().(*Context)

	h, _ := r.Find(GET, "/users/1/files/1", c)
	NotEqual(t, h, nil)
	Equal(t, "1", c.P(0))
	Equal(t, "1", c.P(1))
}

func TestRouterMatchAny(t *testing.T) {
	l := New()
	r := l.router

	// Routes
	r.add(GET, "/", func(*Context) {}, l)
	r.add(GET, "/*", func(*Context) {}, l)
	r.add(GET, "/users/*", func(*Context) {}, l)

	c := l.pool.New().(*Context)

	h, _ := r.Find(GET, "/", c)
	NotEqual(t, h, nil)
	Equal(t, "", c.P(0))

	h, _ = r.Find(GET, "/download", c)
	NotEqual(t, h, nil)
	Equal(t, "download", c.P(0))

	h, _ = r.Find(GET, "/users/joe", c)
	NotEqual(t, h, nil)
	Equal(t, "joe", c.P(0))
}

func TestBadRouter(t *testing.T) {
	l := New()
	l.router = &router{}
	r := l.router

	PanicMatches(t, func() { r.add(GET, "/", func(*Context) {}, l) }, "lars => invalid router initialization")
}

func TestRouterMicroParam(t *testing.T) {
	l := New()
	r := l.router
	r.add(GET, "/:a/:b/:c", func(c *Context) {}, l)

	c := l.pool.New().(*Context)

	h, _ := r.Find(GET, "/1/2/3", c)
	NotEqual(t, h, nil)
	Equal(t, "1", c.P(0))
	Equal(t, "2", c.P(1))
	Equal(t, "3", c.P(2))
}

func TestRouterMixParamMatchAny(t *testing.T) {
	l := New()
	r := l.router

	// Route
	r.add(GET, "/users/:id/*", func(c *Context) {}, l)

	c := l.pool.New().(*Context)

	h, _ := r.Find(GET, "/users/joe/comments", c)
	NotEqual(t, h, nil)

	h(c)
	Equal(t, "joe", c.P(0))
}

func TestRouterMultiRoute(t *testing.T) {
	l := New()
	r := l.router

	fn := func(*Context) {}

	// Routes
	r.add(GET, "/users", fn, l)
	r.add(GET, "/users/:id", fn, l)

	c := l.pool.New().(*Context)

	// Route > /users
	h, _ := r.Find(GET, "/users", c)
	NotEqual(t, h, nil)
	h(c)
	Equal(t, fmt.Sprint(h), fmt.Sprint(fn))

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/1", c)
	NotEqual(t, h, nil)
	Equal(t, fmt.Sprint(h), fmt.Sprint(fn))
	Equal(t, "1", c.P(0))
}

func TestRouterPriority(t *testing.T) {
	l := New()
	r := l.router

	// Routes
	r.add(GET, "/users", func(c *Context) {
		c.Response.Write([]byte("a"))
	}, l)
	r.add(GET, "/users/new", func(c *Context) {
		c.Response.Write([]byte("b"))
	}, l)
	r.add(GET, "/users/:id", func(c *Context) {
		c.Response.Write([]byte(c.P(0)))
	}, l)
	r.add(GET, "/users/dew", func(c *Context) {
		c.Response.Write([]byte("d"))
	}, l)
	r.add(GET, "/users/:id/files", func(c *Context) {
		c.Response.Write([]byte(c.P(0)))
	}, l)
	r.add(GET, "/users/newsee", func(c *Context) {
		c.Response.Write([]byte("f"))
	}, l)
	r.add(GET, "/users/*", func(c *Context) {
		c.Response.Write([]byte(c.Param("_*")))
	}, l)

	c, s := request(GET, "/users", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "a")

	c, s = request(GET, "/users/new", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "b")

	c, s = request(GET, "/users/1", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "1")

	c, s = request(GET, "/users/2/files", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "2")

	c, s = request(GET, "/users/news", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "news")

	c, s = request(GET, "/users/joe/books", l)
	Equal(t, c, http.StatusOK)
	Equal(t, s, "joe/books")
}

func TestRouterParamNames(t *testing.T) {
	l := New()
	r := l.router

	// Routes
	r.add(GET, "/users", func(c *Context) {
		c.Response.Write([]byte("/users"))
	}, l)
	r.add(GET, "/users/:id", func(c *Context) {}, l)
	r.add(GET, "/users/:uid/files/:fid", func(c *Context) {}, l)

	c := l.pool.New().(*Context)

	// Route > /users
	rc, s := request(GET, "/users", l)
	Equal(t, rc, http.StatusOK)
	Equal(t, s, "/users")

	// Route > /users/:id
	h, _ := r.Find(GET, "/users/1", c)
	NotEqual(t, h, nil)
	Equal(t, "1", c.Param("id"))
	Equal(t, "1", c.P(0))

	// Route > /users/:uid/files/:fid
	h, _ = r.Find(GET, "/users/1/files/2", c)
	NotEqual(t, h, nil)
	Equal(t, "1", c.Param("uid"))
	Equal(t, "1", c.P(0))
	Equal(t, "2", c.Param("fid"))
	Equal(t, "2", c.P(1))
}

func TestRouterAPI(t *testing.T) {
	l := New()
	r := l.router

	for _, route := range api {
		r.add(route.Method, route.Path, func(c *Context) {}, l)
	}

	c := l.pool.New().(*Context)

	for _, route := range api {

		h, _ := r.Find(route.Method, route.Path, c)
		NotEqual(t, h, nil)

		for i, n := range c.Params() {
			NotEqual(t, n, "")
			Equal(t, ":"+n, c.P(i))
		}
		h(c)
	}
}

func TestRouterServeHTTP(t *testing.T) {
	r := New()
	r.Get("/users", func(*Context) {})

	// OK
	req, _ := http.NewRequest(GET, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	Equal(t, http.StatusOK, w.Code)

	// Not found
	req, _ = http.NewRequest(GET, "/files", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	Equal(t, http.StatusNotFound, w.Code)
}

// func (n *node) printTree(pfx string, tail bool) {
// 	p := prefix(tail, pfx, "└── ", "├── ")
// 	fmt.Printf("%s%s, %p: type=%d, parent=%p, handler=%v\n", p, n.prefix, n, n.kind, n.parent, n.methodHandler)

// 	children := n.children
// 	l := len(children)
// 	p = prefix(tail, pfx, "    ", "│   ")
// 	for i := 0; i < l-1; i++ {
// 		children[i].printTree(p, false)
// 	}
// 	if l > 0 {
// 		children[l-1].printTree(p, true)
// 	}
// }

func prefix(tail bool, p, on, off string) string {
	if tail {
		return fmt.Sprintf("%s%s", p, on)
	}
	return fmt.Sprintf("%s%s", p, off)
}
