package lars

type router struct {
	tree   *node
	routes []route
	lars   *LARS
}

type node struct {
	kind          kind
	label         byte
	prefix        string
	parent        *node
	children      children
	ppath         string
	pnames        []string
	methodHandler *methodHandler
	lars          *LARS
}

type kind uint8
type children []*node
type methodHandler struct {
	connect HandlerFunc
	delete  HandlerFunc
	get     HandlerFunc
	head    HandlerFunc
	options HandlerFunc
	patch   HandlerFunc
	post    HandlerFunc
	put     HandlerFunc
	trace   HandlerFunc
}

const (
	skind kind = iota
	pkind
	mkind
)

// newRouter returns a new *router instance
func newRouter(l *LARS) *router {
	return &router{
		tree: &node{
			methodHandler: new(methodHandler),
		},
		routes: []route{},
		lars:   l,
	}
}

func (r *router) add(method, path string, h HandlerFunc, l *LARS) {
	ppath := path        // Pristine path
	pnames := []string{} // Param names

	for i, k := 0, len(path); i < k; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil, l)
			for ; i < k && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, k = j, len(path)

			if i == k {
				r.insert(method, path[:i], h, pkind, ppath, pnames, l)
				return
			}
			r.insert(method, path[:i], nil, pkind, ppath, pnames, l)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, "", nil, l)
			pnames = append(pnames, "_*")
			r.insert(method, path[:i+1], h, mkind, ppath, pnames, l)
			return
		}
	}

	r.insert(method, path, h, skind, ppath, pnames, l)
}

func (r *router) insert(method, path string, h HandlerFunc, t kind, ppath string, pnames []string, l *LARS) {
	// Adjust max param
	j := len(pnames)
	if *l.maxParam < j {
		*l.maxParam = j
	}

	cn := r.tree // Current node as root
	if cn == nil {
		panic("lars => invalid router initialization")
	}
	search := path

	for {
		sl := len(search)
		pl := len(cn.prefix)
		j := 0

		// LCP
		max := pl
		if sl < max {
			max = sl
		}
		for ; j < max && search[j] == cn.prefix[j]; j++ {
		}

		if j == 0 {
			// At root node
			cn.label = search[0]
			cn.prefix = search
			if h != nil {
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
				cn.lars = l
			}
		} else if j < pl {
			// Split node
			n := newNode(cn.kind, cn.prefix[j:], cn, cn.children, cn.methodHandler, cn.ppath, cn.pnames, cn.lars)

			// Reset parent node
			cn.kind = skind
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:j]
			cn.children = nil
			cn.methodHandler = new(methodHandler)
			cn.ppath = ""
			cn.pnames = nil
			cn.lars = nil

			cn.addChild(n)

			if j == sl {
				// At parent node
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
				cn.lars = l
			} else {
				// Create child node
				n = newNode(t, search[j:], cn, nil, new(methodHandler), ppath, pnames, l)
				n.addHandler(method, h)
				cn.addChild(n)
			}
		} else if j < sl {
			search = search[j:]
			c := cn.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				cn = c
				continue
			}
			// Create child node
			n := newNode(t, search, cn, nil, new(methodHandler), ppath, pnames, l)
			n.addHandler(method, h)
			cn.addChild(n)
		} else {
			// Node already exists
			if h != nil {
				cn.addHandler(method, h)
				cn.ppath = path
				cn.pnames = pnames
				cn.lars = l
			}
		}
		return
	}
}

func newNode(t kind, pre string, p *node, c children, mh *methodHandler, ppath string, pnames []string, l *LARS) *node {
	return &node{
		kind:          t,
		label:         pre[0],
		prefix:        pre,
		parent:        p,
		children:      c,
		ppath:         ppath,
		pnames:        pnames,
		methodHandler: mh,
		lars:          l,
	}
}

func (n *node) addChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) findChild(l byte, t kind) *node {
	for _, c := range n.children {
		if c.label == l && c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.children {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildByKind(t kind) *node {
	for _, c := range n.children {
		if c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) addHandler(method string, h HandlerFunc) {
	switch method {
	case GET:
		n.methodHandler.get = h
	case POST:
		n.methodHandler.post = h
	case PUT:
		n.methodHandler.put = h
	case DELETE:
		n.methodHandler.delete = h
	case PATCH:
		n.methodHandler.patch = h
	case OPTIONS:
		n.methodHandler.options = h
	case HEAD:
		n.methodHandler.head = h
	case CONNECT:
		n.methodHandler.connect = h
	case TRACE:
		n.methodHandler.trace = h
	}
}

func (n *node) findHandler(method string) HandlerFunc {
	switch method {
	case GET:
		return n.methodHandler.get
	case POST:
		return n.methodHandler.post
	case PUT:
		return n.methodHandler.put
	case DELETE:
		return n.methodHandler.delete
	case PATCH:
		return n.methodHandler.patch
	case OPTIONS:
		return n.methodHandler.options
	case HEAD:
		return n.methodHandler.head
	case CONNECT:
		return n.methodHandler.connect
	case TRACE:
		return n.methodHandler.trace
	default:
		return nil
	}
}

func (n *node) check405(l *LARS) HandlerFunc {
	for _, m := range methods {
		if h := n.findHandler(m); h != nil {
			return methodNotAllowedHandler
		}
	}
	return l.http404
}

func (r *router) Find(method, path string, ctx *Context) (h HandlerFunc, l *LARS) {
	h = r.lars.http404
	l = r.lars
	cn := r.tree // Current node as root

	var (
		search = path
		c      *node  // Child node
		n      int    // Param counter
		nk     kind   // Next kind
		nn     *node  // Next node
		ns     string // Next search
	)

	// Search order static > param > match-any
	for {
		if search == "" {
			goto End
		}

		pl := 0 // Prefix length
		i := 0  // LCP length

		if cn.label != ':' {
			sl := len(search)
			pl = len(cn.prefix)

			// LCP
			max := pl
			if sl < max {
				max = sl
			}
			for ; i < max && search[i] == cn.prefix[i]; i++ {
			}
		}

		if i == pl {
			// Continue search
			search = search[i:]
		} else {
			cn = nn
			search = ns
			if nk == pkind {
				goto Param
			} else if nk == mkind {
				goto MatchAny
			} else {
				// Not found
				return
			}
		}

		if search == "" {
			goto End
		}

		// Static node
		c = cn.findChild(search[0], skind)
		if c != nil {
			// Save next
			if cn.label == '/' {
				nk = pkind
				nn = cn
				ns = search
			}
			cn = c
			continue
		}

		// Param node
	Param:
		c = cn.findChildByKind(pkind)
		if c != nil {
			// Save next
			if cn.label == '/' {
				nk = mkind
				nn = cn
				ns = search
			}
			cn = c
			i, j := 0, len(search)
			for ; i < j && search[i] != '/'; i++ {
			}
			ctx.pvalues[n] = search[:i]
			n++
			search = search[i:]
			continue
		}

		// Match-any node
	MatchAny:
		if cn = cn.findChildByKind(mkind); cn == nil {
			// Not found
			return
		}
		ctx.pvalues[len(cn.pnames)-1] = search
		goto End
	}

End:

	ctx.path = cn.ppath
	ctx.pnames = cn.pnames
	h = cn.findHandler(method)

	if cn.lars != nil {
		l = cn.lars
	}

	// NOTE: Slow zone...
	if h == nil {

		h = cn.check405(l.lars)

		// Dig further for match-any, might have an empty value for *, e.g.
		if cn = cn.findChildByKind(mkind); cn == nil {
			return
		}

		ctx.pvalues[len(cn.pnames)-1] = ""

		if h = cn.findHandler(method); h == nil {
			h = cn.check405(l.lars)
		}
	}

	return
}
