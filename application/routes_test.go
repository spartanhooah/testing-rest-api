package application

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
	"testing"
)

func Test_application_routes(t *testing.T) {
	var registered = []struct {
		route  string
		method string
	}{
		{"/auth", "POST"},
		{"/refresh-token", "POST"},
		{"/users/", "GET"},
		{"/users/{userId}", "GET"},
		{"/users/{userId}", "DELETE"},
		{"/users/", "PUT"},
		{"/users/", "PATCH"},
	}

	mux := app.Routes()

	chiRoutes := mux.(chi.Routes)

	for _, route := range registered {
		// check if the route exists
		if !routeExists(route.route, route.method, chiRoutes) {
			t.Errorf("route %s does not exist for method %s", route.route, route.method)
		}
	}
}

func routeExists(testRoute, testMethod string, chiRoutes chi.Routes) bool {
	found := false

	_ = chi.Walk(chiRoutes, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if strings.EqualFold(method, testMethod) && strings.EqualFold(route, testRoute) {
			found = true
		}

		return nil
	})

	return found
}
