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

const getOrganizationIDResponseTemplate = `{
    "data": {
        "organization": {
            "id": "%s"
        }
    }
}`

const getOrganizationIPAllowListEntriesResponseTemplate = `{
    "data": {
        "organization": {
            "ipAllowListEntries": {
                "nodes": [
                    {
                        "allowListValue": "%s",
                        "isActive": %t,
                        "name": "%s",
                        "id": "%s",
                        "createdAt": "%s",
                        "updatedAt": "%s"
                    }
                ],
                "pageInfo": {
                    "hasNextPage": %t,
                    "startCursor": "abc",
                    "endCursor": "abc"
                }
            }
        }
    }
}`

func TestGetOrganizationID(t *testing.T) {
	// given
	expectedOrganizationID := "abc123"
	gitHubGraphQLAPIMock := serverReturning(getOrganizationIDResponseWith(expectedOrganizationID))
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	retrievedOrganizationID, err := client.GetOrganizationID(context.TODO(), "some organization")

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedOrganizationID, retrievedOrganizationID)
}

func TestGetOrganizationIDWithFailingServer(t *testing.T) {
	// given
	gitHubGraphQLAPIMock := serverReturningAnEmptyResponseWith(http.StatusInternalServerError)
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	retrievedOrganizationID, err := client.GetOrganizationID(context.TODO(), "some organization")

	// then
	assert.Error(t, err)
	assert.Equal(t, retrievedOrganizationID, "")
}

func TestGetOrganizationIPAllowListEntriesWithPagedResponseOneEntryPerPage(t *testing.T) {
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
			gitHubGraphQLAPIMock := serverWithOneEntryPerPageGetOrganizationIPAllowListEntries(test.expectedEntries)
			client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

			// when
			entries, err := client.GetOrganizationIPAllowListEntries(context.TODO(), "some organization")

			// then
			assert.NoError(t, err)
			assert.Len(t, entries, len(test.expectedEntries))
			assert.Equal(t, test.expectedEntries, entries)
		})
	}
}

func TestGetOrganizationIPAllowListEntriesWithEntriesCachingCallsAPIOnlyOnce(t *testing.T) {
	// given
	expectedEntry := IPAllowListEntry{
		ID:             "some-id",
		CreatedAt:      truncateToGitHubPrecision(time.Now()),
		UpdatedAt:      truncateToGitHubPrecision(time.Now()),
		AllowListValue: "1.2.3.4/32",
		IsActive:       true,
		Name:           "Managed by Terraform",
	}
	gitHubGraphQLAPIMock, receivedRequests := serverReturningConsecutiveResponses(getOrganizationIPAllowListEntriesResponseLastPageWith(expectedEntry))
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	_, _ = client.GetOrganizationIPAllowListEntries(context.TODO(), "some organization")
	entries, err := client.GetOrganizationIPAllowListEntries(context.TODO(), "some organization")

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedEntry, *entries[0])
	assert.Equal(t, int64(1), receivedRequests.Load())
}

func TestGetOrganizationIPAllowListEntriesWithEntriesCachingDoesNotExecuteCallsConcurrently(t *testing.T) {
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
			gitHubGraphQLAPIMock, receivedRequests := serverWaitingWithWritingAResponseUntilAllRequestsAreReceived(2)
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

func getOrganizationIDResponseWith(expectedOrganizationID string) string {
	return fmt.Sprintf(getOrganizationIDResponseTemplate, expectedOrganizationID)
}

func getOrganizationIPAllowListEntriesResponseLastPageWith(expectedEntry IPAllowListEntry) string {
	hasNextPage := false
	return fmt.Sprintf(getOrganizationIPAllowListEntriesResponseTemplate, expectedEntry.AllowListValue, expectedEntry.IsActive, expectedEntry.Name, expectedEntry.ID, expectedEntry.CreatedAt.Format(gitHubTimeFormat), expectedEntry.UpdatedAt.Format(gitHubTimeFormat), hasNextPage)
}

func getOrganizationIPAllowListEntriesResponseWithNextPageAnd(expectedEntry IPAllowListEntry) string {
	hasNextPage := true
	return fmt.Sprintf(getOrganizationIPAllowListEntriesResponseTemplate, expectedEntry.AllowListValue, expectedEntry.IsActive, expectedEntry.Name, expectedEntry.ID, expectedEntry.CreatedAt.Format(gitHubTimeFormat), expectedEntry.UpdatedAt.Format(gitHubTimeFormat), hasNextPage)
}

func serverWithOneEntryPerPageGetOrganizationIPAllowListEntries(expectedEntries []*IPAllowListEntry) *httptest.Server {
	gitHubGraphQLAPIMock, _ := serverReturningConsecutiveResponses(pagedGetOrganizationIPAllowListEntriesResponses(expectedEntries)...)
	return gitHubGraphQLAPIMock
}

func pagedGetOrganizationIPAllowListEntriesResponses(expectedEntries []*IPAllowListEntry) []string {
	pagedResponses := make([]string, len(expectedEntries)-1, len(expectedEntries))
	for i := range pagedResponses {
		pagedResponses[i] = getOrganizationIPAllowListEntriesResponseWithNextPageAnd(*expectedEntries[i])
	}
	lastPage := getOrganizationIPAllowListEntriesResponseLastPageWith(*expectedEntries[len(expectedEntries)-1])
	return append(pagedResponses, lastPage)
}
