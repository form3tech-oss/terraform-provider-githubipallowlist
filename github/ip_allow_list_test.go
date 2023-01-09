package github

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

const createEntryResponseTemplate = `
{
    "data": {
        "createIpAllowListEntry": {
            "ipAllowListEntry": {
                "id": "%s",
                "createdAt": "%s",
                "updatedAt": "%s",
                "allowListValue": "%s",
                "isActive": %t,
                "name": "%s"
            }
        }
    }
}`

const deleteEntryResponseTemplate = `
{
    "data": {
        "deleteIpAllowListEntry": {
            "ipAllowListEntry": {
                "id": "%s"
            }
        }
    }
}`

const deleteEntryResponseForMissingEntry = `
{
    "data": {
        "deleteIpAllowListEntry": null
    },
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
            "message": "Could not resolve to a node with the global id of 'abc-123'."
        }
    ]
}`

func TestCreateIPAllowListEntry(t *testing.T) {
	// given
	expectedEntry := IPAllowListEntry{
		ID:             "some id",
		AllowListValue: "1.2.3.4/32",
		IsActive:       true,
		Name:           "some name",
		CreatedAt:      time.Now().UTC().Truncate(time.Second),
		UpdatedAt:      time.Now().UTC().Truncate(time.Second),
	}
	gitHubGraphQLAPIMock := serverReturning(createEntryResponseWith(expectedEntry))
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	createdEntry, err := client.CreateIPAllowListEntry(context.TODO(), "some owner", expectedEntry.Name, expectedEntry.AllowListValue, expectedEntry.IsActive)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedEntry, *createdEntry)
}

func TestCreateIPAllowListEntryWithFailingServer(t *testing.T) {
	// given
	gitHubGraphQLAPIMock := serverReturningAnEmptyResponseWith(http.StatusInternalServerError)
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	createdEntry, err := client.CreateIPAllowListEntry(context.TODO(), "some owner", "some name", "some value", true)

	// then
	assert.Error(t, err)
	assert.Nil(t, createdEntry)
}

func TestDeleteIPAllowListEntry(t *testing.T) {
	// given
	expectedEntryID := "expected-entry-id"
	gitHubGraphQLAPIMock := serverReturning(deleteEntryResponseWith(expectedEntryID))
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	deletedEntryID, err := client.DeleteIPAllowListEntry(context.TODO(), expectedEntryID)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedEntryID, deletedEntryID)
}

func TestDeleteIPAllowListEntryWithMissingEntry(t *testing.T) {
	// given
	gitHubGraphQLAPIMock := serverReturning(deleteEntryResponseForMissingEntry)
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	deletedEntryID, err := client.DeleteIPAllowListEntry(context.TODO(), "some-entry-id")

	// then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Could not resolve to a node with the global id of 'abc-123'.")
	assert.Empty(t, deletedEntryID)
}

func TestDeleteIPAllowListEntryWithFailingServer(t *testing.T) {
	// given
	expectedStatusCode := http.StatusInternalServerError
	gitHubGraphQLAPIMock := serverReturningAnEmptyResponseWith(expectedStatusCode)
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	deletedEntryID, err := client.DeleteIPAllowListEntry(context.TODO(), "some-entry-id")

	// then
	var target ErrorWithStatusCode
	assert.ErrorAs(t, err, &target)
	assert.Equal(t, target.StatusCode, expectedStatusCode)
	assert.Empty(t, deletedEntryID)
}

func createEntryResponseWith(expectedEntry IPAllowListEntry) string {
	res := fmt.Sprintf(createEntryResponseTemplate, expectedEntry.ID, expectedEntry.CreatedAt.Format(gitHubTimeFormat), expectedEntry.UpdatedAt.Format(gitHubTimeFormat), expectedEntry.AllowListValue, expectedEntry.IsActive, expectedEntry.Name)
	return res
}

func deleteEntryResponseWith(expectedEntryID string) string {
	return fmt.Sprintf(deleteEntryResponseTemplate, expectedEntryID)
}
