package github

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const getEnterpriseIDResponseTemplate = `{
    "data": {
        "enterprise": {
            "id": "%s"
        }
    }
}`

const getEnterpriseIPAllowListEntriesResponseTemplate = `{
    "data": {
        "enterprise": {
            "ownerInfo": {
                "ipAllowListEntries": {
                    "nodes": [
                        {
                            "id": "%s",
                            "allowListValue": "%s",
                            "name": "%s",
                            "isActive": %t,
                            "createdAt": "%s",
                            "updatedAt": "%s"
                        }
                    ],
                    "pageInfo": {
                        "hasNextPage": %t,
                        "startCursor": "abc-123",
                        "endCursor": "abc-123"
                    }
                }
            }
        }
    }
}`

const someGetEnterpriseIPAllowListEntriesResponse = `{
    "data": {
        "enterprise": {
            "ownerInfo": {
                "ipAllowListEntries": {
                    "nodes": [
                        {
                            "id": "some-id",
                            "allowListValue": "1.2.3.4/32",
                            "name": null,
                            "isActive": false,
                            "createdAt": "2023-01-01T11:11:11Z",
                            "updatedAt": "2023-01-01T11:11:11Z"
                        }
                    ],
                    "pageInfo": {
                        "hasNextPage": true,
                        "startCursor": "abc-123",
                        "endCursor": "abc-123"
                    }
                }
            }
        }
    }
}`

func TestGetEnterpriseID(t *testing.T) {
	// given
	expectedEnterpriseID := "abc123"
	gitHubGraphQLAPIMock := serverReturning(getEnterpriseIDResponseWith(expectedEnterpriseID))
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	retrievedEnterpriseID, err := client.GetEnterpriseID(context.TODO(), "some enterprise")

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedEnterpriseID, retrievedEnterpriseID)
}

func TestGetEnterpriseIDWithFailingServer(t *testing.T) {
	// given
	expectedStatusCode := http.StatusInternalServerError
	gitHubGraphQLAPIMock := serverReturningAnEmptyResponseWith(expectedStatusCode)
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	retrievedEnterpriseID, err := client.GetEnterpriseID(context.TODO(), "some enterprise")

	// then
	var target ErrorWithStatusCode
	assert.ErrorAs(t, err, &target)
	assert.Equal(t, expectedStatusCode, target.StatusCode)
	assert.Equal(t, retrievedEnterpriseID, "")
}

func TestGetEnterpriseIPAllowListEntriesWithPagedResponseOneEntryPerPage(t *testing.T) {
	tests := []struct {
		expectedEntries []*IPAllowListEntry
	}{
		{
			expectedEntries: []*IPAllowListEntry{
				{
					ID:             "some-id",
					CreatedAt:      truncateToGitHubPrecision(time.Now()),
					UpdatedAt:      truncateToGitHubPrecision(time.Now()),
					AllowListValue: "1.2.3.4/32",
					IsActive:       true,
					Name:           "Managed by Terraform",
				}},
		},
		{
			expectedEntries: []*IPAllowListEntry{
				{
					ID:             "some-id1",
					CreatedAt:      truncateToGitHubPrecision(time.Now()),
					UpdatedAt:      truncateToGitHubPrecision(time.Now()),
					AllowListValue: "1.2.3.4/32",
					IsActive:       false,
					Name:           "Managed by Terraform",
				},
				{
					ID:             "some-id2",
					CreatedAt:      truncateToGitHubPrecision(time.Now()),
					UpdatedAt:      truncateToGitHubPrecision(time.Now()),
					AllowListValue: "1.2.3.5/32",
					IsActive:       true,
					Name:           "Managed by Terraform",
				}},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("number of pages:%d", len(test.expectedEntries)), func(t *testing.T) {
			// given
			gitHubGraphQLAPIMock := serverWithOneEntryPerPageGetEnterpriseIPAllowListEntries(test.expectedEntries)
			client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

			// when
			entries, err := client.getEnterpriseIPAllowListEntries(context.TODO(), "some enterprise")

			// then
			assert.NoError(t, err)
			assert.Len(t, entries, len(test.expectedEntries))
			assert.Equal(t, test.expectedEntries, entries)
		})
	}
}

func TestGetEnterpriseIPAllowListEntriesWithEntriesCachingCallsAPIOnlyOnce(t *testing.T) {
	// given
	expectedEntry := IPAllowListEntry{
		ID:             "some-id",
		CreatedAt:      truncateToGitHubPrecision(time.Now()),
		UpdatedAt:      truncateToGitHubPrecision(time.Now()),
		AllowListValue: "1.2.3.4/32",
		IsActive:       true,
		Name:           "Managed by Terraform",
	}
	gitHubGraphQLAPIMock, receivedRequests := serverReturningConsecutiveResponses(getEnterpriseIPAllowListEntriesResponseLastPageWith(expectedEntry))
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	_, _ = client.GetEnterpriseIPAllowListEntries(context.TODO(), "some enterprise")
	entries, err := client.GetEnterpriseIPAllowListEntries(context.TODO(), "some enterprise")

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedEntry, *entries[0])
	assert.Equal(t, int64(1), receivedRequests.Load())
}

func TestGetEnterpriseIPAllowListEntriesWithEntriesCachingDoesNotExecuteCallsConcurrently(t *testing.T) {
	tests := []struct {
		concurrentRequests int
	}{
		{1},
		{2},
		{10},
		{100},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Concurrency:%d", test.concurrentRequests), func(t *testing.T) {
			// given
			gitHubGraphQLAPIMock, receivedRequests := serverWaitingWithWritingAResponseUntilAllRequestsAreReceived(2, someGetEnterpriseIPAllowListEntriesResponse)
			client := NewAuthenticatedGitHubClient(context.TODO(), "", WithConcurrency(int64(test.concurrentRequests)), WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

			// when
			for i := 0; i < test.concurrentRequests; i++ {
				go func() {
					_, _ = client.GetEnterpriseIPAllowListEntries(context.TODO(), "some enterprise")
				}()
			}

			// then
			withTimeout(
				receivedRequests.Wait,
				func() {
					assert.Fail(t, "Client called the API concurrently. Received too many requests which suggests an error in client's cache concurrency control.")
				},
				func() {
					receivedRequests.Done()
				},
				100*time.Millisecond,
			)
		})
	}
}

func getEnterpriseIDResponseWith(expectedEnterpriseID string) string {
	return fmt.Sprintf(getEnterpriseIDResponseTemplate, expectedEnterpriseID)
}

func serverWithOneEntryPerPageGetEnterpriseIPAllowListEntries(expectedEntries []*IPAllowListEntry) *httptest.Server {
	gitHubGraphQLAPIMock, _ := serverReturningConsecutiveResponses(pagedGetEnterpriseIPAllowListEntriesResponses(expectedEntries)...)
	return gitHubGraphQLAPIMock
}

func pagedGetEnterpriseIPAllowListEntriesResponses(expectedEntries []*IPAllowListEntry) []string {
	pagedResponses := make([]string, len(expectedEntries)-1, len(expectedEntries))
	for i := range pagedResponses {
		pagedResponses[i] = getEnterpriseIPAllowListEntriesResponseWithNextPageAnd(*expectedEntries[i])
	}
	lastPage := getEnterpriseIPAllowListEntriesResponseLastPageWith(*expectedEntries[len(expectedEntries)-1])
	return append(pagedResponses, lastPage)
}

func getEnterpriseIPAllowListEntriesResponseLastPageWith(expectedEntry IPAllowListEntry) string {
	hasNextPage := false
	return fmt.Sprintf(getEnterpriseIPAllowListEntriesResponseTemplate, expectedEntry.ID, expectedEntry.AllowListValue, expectedEntry.Name, expectedEntry.IsActive, expectedEntry.CreatedAt.Format(gitHubTimeFormat), expectedEntry.UpdatedAt.Format(gitHubTimeFormat), hasNextPage)
}

func getEnterpriseIPAllowListEntriesResponseWithNextPageAnd(expectedEntry IPAllowListEntry) string {
	hasNextPage := true
	return fmt.Sprintf(getEnterpriseIPAllowListEntriesResponseTemplate, expectedEntry.ID, expectedEntry.AllowListValue, expectedEntry.Name, expectedEntry.IsActive, expectedEntry.CreatedAt.Format(gitHubTimeFormat), expectedEntry.UpdatedAt.Format(gitHubTimeFormat), hasNextPage)
}
