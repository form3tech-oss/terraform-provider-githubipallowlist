package github

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	someGetOrganizationIPAllowListEntriesResponse = `{
    "data": {
        "organization": {
            "ipAllowListEntries": {
                "nodes": [
                    {
                        "allowListValue": "1.1.1.1",
                        "isActive": true,
                        "name": null,
                        "id": "abc"
                    }
                ],
                "pageInfo": {
                    "hasNextPage": false,
                    "startCursor": "abc",
                    "endCursor": "abc"
                }
            }
        }
    }
}`

	pagedGetOrganizationIPAllowListEntriesResponse = `{
    "data": {
        "organization": {
            "ipAllowListEntries": {
                "nodes": [
                    {
                        "allowListValue": "1.1.1.1",
                        "isActive": true,
                        "name": null,
                        "id": "abc"
                    }
                ],
                "pageInfo": {
                    "hasNextPage": true,
                    "startCursor": "abc",
                    "endCursor": "abc"
                }
            }
        }
    }
}`
	responseWithGraphQLErrorTemplate = `{
    "data": null,
    "errors": [
        {
            "type": "NOT_FOUND",
            "path": [
                "deleteIpAllowListEntry"
            ],
            "locations": [
                {
                    "line": 2,
                    "column": 3
                }
            ],
            "message": "%s"
        }
    ]
}`
)

func TestNewGitHubClient(t *testing.T) {
	client := NewAuthenticatedGitHubClient(context.TODO(), "")
	assert.NotNil(t, client)
}

func TestClientGetsAllPages(t *testing.T) {
	tests := []struct {
		expectedPages int64
	}{
		{int64(2)},
		{int64(10)},
		{int64(100)},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Expected pages:%d", test.expectedPages), func(t *testing.T) {
			// given
			gitHubGraphQLAPIMock, noRequestsServed := serverWithPagedResponse(test.expectedPages)
			client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

			// when
			_, _ = client.GetOrganizationIPAllowListEntries(context.TODO(), "some organisation")

			// then
			assert.Equal(t, test.expectedPages, noRequestsServed.Load())
		})
	}
}

func TestClientHandlesGraphQLErrors(t *testing.T) {
	// given
	expectedErrorMessage := "Could not resolve to a node with the global id of 'abc'."
	gitHubGraphQLAPIMock := serverReturningGrapeQLError(expectedErrorMessage)
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	_, err := client.GetOrganizationIPAllowListEntries(context.TODO(), "some organisation")

	// then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErrorMessage)
}

func TestClientHandlesOtherErrors(t *testing.T) {
	// given
	expectedStatusCode := http.StatusInternalServerError
	gitHubGraphQLAPIMock := serverReturningAnEmptyResponseWith(expectedStatusCode)
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	_, err := client.GetOrganizationIPAllowListEntries(context.TODO(), "some organisation")

	// then
	var target ErrorWithStatusCode
	assert.ErrorAs(t, err, &target)
	assert.Equal(t, expectedStatusCode, target.StatusCode)
}

func TestClientCanExecuteRequestsConcurrently(t *testing.T) {
	tests := []struct {
		concurrentRequests int
	}{
		{1},
		{2},
		{10},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Concurrency:%d", test.concurrentRequests), func(t *testing.T) {
			// given
			gitHubGraphQLAPIMock, receivedRequests := serverWaitingWithWritingAResponseUntilAllRequestsAreReceived(test.concurrentRequests)
			client := NewAuthenticatedGitHubClient(context.TODO(), "", WithConcurrency(int64(test.concurrentRequests)), WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

			// when
			for i := 0; i < test.concurrentRequests; i++ {
				go func() {
					_, _ = client.GetOrganizationIPAllowListEntries(context.TODO(), "some organisation")
				}()
			}

			// then
			withTimeout(
				receivedRequests.Wait,
				func() {},
				func() {
					assert.Fail(t, "Timed out waiting for all requests. Some requests are missing which suggests an error in client's concurrency control.")
				},
				100*time.Millisecond,
			)
		})
	}
}

func TestClientCanNotExceedMaxConcurrentRequests(t *testing.T) {
	tests := []struct {
		concurrentRequests int
	}{
		{1},
		{2},
		{10},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Concurrency:%d", test.concurrentRequests), func(t *testing.T) {
			// given
			gitHubGraphQLAPIMock, receivedRequests := serverWaitingWithWritingAResponseUntilAllRequestsAreReceived(test.concurrentRequests + 1)
			client := NewAuthenticatedGitHubClient(context.TODO(), "", WithConcurrency(int64(test.concurrentRequests)), WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

			// when
			for i := 0; i < test.concurrentRequests; i++ {
				go func() {
					_, _ = client.GetOrganizationIPAllowListEntries(context.TODO(), "some organisation")
				}()
			}

			// then
			withTimeout(
				receivedRequests.Wait,
				func() {
					assert.Fail(t, "Client exceeded concurrency. Received too many requests which suggests an error in client's concurrency control.")
				},
				func() {
					receivedRequests.Done()
				},
				100*time.Millisecond,
			)
		})
	}
}

func withTimeout(awaited func(), onDone func(), onTimeout func(), timeout time.Duration) {
	ticker := time.NewTimer(timeout)
	done := make(chan struct{})
	go func() {
		awaited()
		done <- struct{}{}
	}()
	select {
	case <-ticker.C:
		onTimeout()
	case <-done:
		onDone()
	}
}

func serverWaitingWithWritingAResponseUntilAllRequestsAreReceived(expectedRequests int) (*httptest.Server, *sync.WaitGroup) {
	var receivedRequests sync.WaitGroup
	receivedRequests.Add(expectedRequests)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequests.Done()
		receivedRequests.Wait()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(someGetOrganizationIPAllowListEntriesResponse))
	}))
	return testServer, &receivedRequests
}

func serverWithPagedResponse(numberOfPagesToReturn int64) (*httptest.Server, *atomic.Int64) {
	return serverReturningConsecutiveResponses(generatePagedGetOrganizationIPAllowListEntriesResponses(int(numberOfPagesToReturn))...)
}

func generatePagedGetOrganizationIPAllowListEntriesResponses(numberOfPagesToReturn int) []string {
	pagedResponses := make([]string, numberOfPagesToReturn-1, numberOfPagesToReturn)
	for i := range pagedResponses {
		pagedResponses[i] = pagedGetOrganizationIPAllowListEntriesResponse
	}
	return append(pagedResponses, someGetOrganizationIPAllowListEntriesResponse)
}

func serverReturningGrapeQLError(errorMessage string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(responseWithGraphQLErrorTemplate, errorMessage)))
	}))
}
