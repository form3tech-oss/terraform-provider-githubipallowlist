package github

import (
	"net/http"
	"net/http/httptest"
)

func serverReturning(res string) *httptest.Server {
	gitHubGraphQLAPIMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(res))
	}))
	return gitHubGraphQLAPIMock
}

func serverReturningInternalServerError() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
}
