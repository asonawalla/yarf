package yarf

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockResource struct {
	Resource
}

type MockMiddleware struct {
	Middleware
}

func TestYarfAdd(t *testing.T) {
	y := New()
	r := new(MockResource)

	y.Add("/test", r)

	if len(y.routes) != 1 {
		t.Fatalf("Added 1 route, found %d in the list.", len(y.routes))
	}
	if y.routes[0].(*route).path != "/test" {
		t.Fatalf("Added /test path. Found %s", y.routes[0].(*route).path)
	}
	if y.routes[0].(*route).handler != r {
		t.Fatal("Added a Handler. Handler found seems to be different")
	}

	y.Add("/test/2", r)

	if len(y.routes) != 2 {
		t.Fatalf("Added 2 routes, found %d routes in the list.", len(y.routes))
	}

	if y.routes[0].(*route).handler != y.routes[1].(*route).handler {
		t.Fatal("Added a Handler to 2 routes. Handlers found seems to be different")
	}
}

func TestYarfAddGroup(t *testing.T) {
	y := New()
	g := RouteGroup("/group")

	y.AddGroup(g)

	if len(y.routes) != 1 {
		t.Fatalf("Added 1 route group, found %d in the list.", len(y.routes))
	}
	if y.routes[0].(*routeGroup).prefix != "/group" {
		t.Fatalf("Added a /group route prefix. Found %s", y.routes[0].(*routeGroup).prefix)
	}
}

func TestYarfInsert(t *testing.T) {
	y := New()
	m := new(MockMiddleware)

	y.Insert(m)

	if len(y.middleware) != 1 {
		t.Fatalf("Added 1 middleware, found %d in the list.", len(y.routes))
	}
	if y.middleware[0] != m {
		t.Fatal("Added a middleware. Stored one seems to be different")
	}
}

func TestYarfCache(t *testing.T) {
	y := New()

	if len(y.cache.storage) > 0 {
		t.Error("yarf.cache.storage should be empty after initialization")
	}

	r := new(MockResource)
	y.Add("/test", r)

	req, _ := http.NewRequest("GET", "http://localhost:8080/route/not/match", nil)
	res := httptest.NewRecorder()
	y.ServeHTTP(res, req)

	if len(y.cache.storage) > 0 {
		t.Error("yarf.cache.storage should be empty after non-matching request")
	}

	req, _ = http.NewRequest("GET", "http://localhost:8080/test", nil)
	y.ServeHTTP(res, req)

	if len(y.cache.storage) != 1 {
		t.Error("yarf.cache.storage should have 1 item after matching request")
	}

	for i := 0; i < 100; i++ {
		y.ServeHTTP(res, req)
	}

	if len(y.cache.storage) != 1 {
		t.Error("yarf.cache.storage should have 1 item after multiple matching requests to a single route")
	}
}

func TestYarfUseCacheFalse(t *testing.T) {
	r := new(MockResource)
	y := New()
	y.UseCache = false
	y.Add("/test", r)

	req, _ := http.NewRequest("GET", "http://localhost:8080/test", nil)
	res := httptest.NewRecorder()
	y.ServeHTTP(res, req)

	if len(y.cache.storage) > 0 {
		t.Error("yarf.cache.storage should be empty after matching request with yarf.UseCache = false")
	}
}

func TestRace(t *testing.T) {
	g := RouteGroup("/test")
	g.Add("/one/:param", &MockResource{})
	g.Add("/two/:param", &MockResource{})

	y := New()
	y.AddGroup(g)

	one, _ := http.NewRequest("GET", "http://localhost:8080/test/one/1", nil)
	two, _ := http.NewRequest("GET", "http://localhost:8080/test/two/2", nil)

	for i := 0; i < 1000; i++ {
		res1 := httptest.NewRecorder()
		res2 := httptest.NewRecorder()

		go y.ServeHTTP(res1, one)
		go y.ServeHTTP(res2, two)
	}
}
