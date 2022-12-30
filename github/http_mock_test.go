package github

import (
	"net/http"
	"net/http/httptest"
)

func serverReturning(body string) *httptest.Server {
	gitHubGraphQLAPIMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	return gitHubGraphQLAPIMock
}

func serverReturningAnEmptyResponseWith(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}
