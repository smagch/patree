package treemux

import (
	"net/http"
	"testing"
)

type foobarHandler struct{}

func (h *foobarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func TestStaticEntry(t *testing.T) {
	e := newStatic("/foobar")
	e.handlers["GET"] = &foobarHandler{}

	h, _ := e.Exec("GET", "/foobar")
	if h == nil {
		t.Fatal("static match is broken")
	}

	h, _ = e.Exec("POST", "/foobar")
	if h != nil {
		t.Fatal("cought wrong method")
	}

	h, _ = e.Exec("GET", "/foobar/2000")
	if h != nil {
		t.Fatal("cought wrong path")
	}

	child := newStatic("/2000")
	child.handlers["GET"] = &foobarHandler{}
	e.add(child)
	h, _ = e.Exec("GET", "/foobar/2000")
	if h == nil {
		t.Fatal("nested entry match failed")
	}

	parent := newStatic("/api")
	parent.add(e)
	h, _ = parent.Exec("GET", "/api/foobar/2000")
	if h == nil {
		t.Fatal("should catch nested entry")
	}
	h, _ = parent.Exec("GET", "/api/foobar")
	if h == nil {
		t.Fatal("should catch nested entry")
	}
	h, _ = parent.Exec("GET", "/api")
	if h != nil {
		t.Fatal("should not catch nested entry")
	}
}
