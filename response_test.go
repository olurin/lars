package lars

import (
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

func TestResponse(t *testing.T) {
	w := httptest.NewRecorder()
	r := &Response{ResponseWriter: w}

	// Header
	NotEqual(t, r.Header(), nil)

	// WriteHeader
	r.WriteHeader(http.StatusOK)
	Equal(t, http.StatusOK, r.Status())

	// Committed
	Equal(t, r.Committed(), true)

	// Already committed
	r.WriteHeader(http.StatusTeapot)
	NotEqual(t, http.StatusTeapot, r.Status())

	// Status
	r.status = http.StatusOK
	Equal(t, http.StatusOK, r.Status())

	// Write
	s := "l"
	n, err := r.Write([]byte(s))
	Equal(t, err, nil)
	Equal(t, n, 1)

	// Size
	Equal(t, int64(len(s)), r.Size())

	// WriteString
	s = "lars"
	n, err = r.WriteString(s)
	Equal(t, err, nil)
	Equal(t, n, 4)

	// Flush
	r.Flush()

	// Size
	Equal(t, int64(len(s))+1, r.Size())

	// Committed
	Equal(t, true, r.Committed())

	// Hijack
	PanicMatches(t, func() { r.Hijack() }, "interface conversion: *httptest.ResponseRecorder is not http.Hijacker: missing method Hijack")

	// CloseNotify
	PanicMatches(t, func() { r.CloseNotify() }, "interface conversion: *httptest.ResponseRecorder is not http.CloseNotifier: missing method CloseNotify")

	// reset
	r.reset(httptest.NewRecorder(), nil)
}
