package application

import "net/http"

func (app *Application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Access-Control-Allow-Origin", "http://localhost:8090")

		if req.Method == "OPTIONS" {
			resp.Header().Set("Access-Control-Allow-Credentials", "true")
			resp.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			resp.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")

			return
		}

		next.ServeHTTP(resp, req)
	})
}

func (app *Application) authRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		_, _, err := app.getTokenFromHeaderAndVerify(resp, req)

		if err != nil {
			resp.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(resp, req)
	})
}
