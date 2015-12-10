package lars

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
)

type Response struct {
	http.ResponseWriter
	status    int
	size      int64
	committed bool
	lars      *LARS
}

func (r *Response) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *Response) Writer() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *Response) WriteHeader(code int) {
	if r.committed {
		log.Println("response already committed")
		return
	}
	r.status = code
	r.ResponseWriter.WriteHeader(code)
	r.committed = true
}

func (r *Response) Write(b []byte) (n int, err error) {
	n, err = r.ResponseWriter.Write(b)
	r.size += int64(n)
	return n, err
}

// WriteString write string to ResponseWriter
func (r *Response) WriteString(s string) (n int, err error) {
	n, err = io.WriteString(r.ResponseWriter, s)
	r.size += int64(n)
	return
}

// Flush wraps response writer's Flush function.
func (r *Response) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

// Hijack wraps response writer's Hijack function.
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify wraps response writer's CloseNotify function.
func (r *Response) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Size() int64 {
	return r.size
}

func (r *Response) Committed() bool {
	return r.committed
}

func (r *Response) reset(w http.ResponseWriter, e *LARS) {
	r.ResponseWriter = w
	r.size = 0
	r.status = http.StatusOK
	r.committed = false
	r.lars = e
}
