package github

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

const (
	gitHubTimeFormat = "2006-01-02T15:04:05Z"
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

func truncateToGitHubPrecision(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
}

func serverReturningConsecutiveResponses(responseBodies ...string) (*httptest.Server, *atomic.Int64) {
	var requestSent atomic.Int64
	requestSent.Store(0)
	gitHubGraphQLAPIMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentRequest := requestSent.Load()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBodies[currentRequest]))
		requestSent.Add(1)
	}))
	return gitHubGraphQLAPIMock, &requestSent
}
