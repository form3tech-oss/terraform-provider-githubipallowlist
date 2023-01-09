package github

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var someIpAllowListEntryParameters = IPAllowListEntryParameters{
	Name:     "some-name",
	Value:    "some value",
	IsActive: false,
}

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

const updateEntryResponseTemplate = `
{
    "data": {
        "updateIpAllowListEntry": {
            "ipAllowListEntry": {
                "id": "%s",
                "allowListValue": "%s",
                "isActive": %t,
                "name": "%s",
                "createdAt": "%s",
                "updatedAt": "%s"
            }
        }
    }
}`

const updateEntryResponseForMissingEntry = `
{
    "data": {
        "updateIpAllowListEntry": null
    },
    "errors": [
        {
            "type": "NOT_FOUND",
            "path": [
                "updateIpAllowListEntry"
            ],
            "locations": [
                {
                    "line": 2,
                    "column": 3
                }
            ],
            "message": "Could not resolve to a node with the global id of 'abc-123'"
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

func TestUpdateIPAllowListEntry(t *testing.T) {
	// given
	expectedEntry := IPAllowListEntry{
		ID:             "some-entry-id",
		CreatedAt:      time.Now().UTC().Truncate(time.Second),
		UpdatedAt:      time.Now().UTC().Truncate(time.Second),
		AllowListValue: "1.2.3.4/32",
		IsActive:       true,
		Name:           "some name",
	}
	expectedParameters := IPAllowListEntryParameters{
		Name:     expectedEntry.Name,
		Value:    expectedEntry.AllowListValue,
		IsActive: expectedEntry.IsActive,
	}
	gitHubGraphQLAPIMock := serverReturning(updateEntryResponseWith(expectedEntry))
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	updatedEntry, err := client.UpdateIPAllowListEntry(context.TODO(), "some-entry-id", expectedParameters)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedEntry, *updatedEntry)
}

func TestUpdateIPAllowListEntryWithMissingEntry(t *testing.T) {
	// given
	gitHubGraphQLAPIMock := serverReturning(updateEntryResponseForMissingEntry)
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	deletedEntryID, err := client.UpdateIPAllowListEntry(context.TODO(), "some-entry-id", someIpAllowListEntryParameters)

	// then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Could not resolve to a node with the global id of 'abc-123'")
	assert.Empty(t, deletedEntryID)
}

func TestUpdateIPAllowListEntryWithFailingServer(t *testing.T) {
	// given
	expectedStatusCode := http.StatusInternalServerError
	gitHubGraphQLAPIMock := serverReturningAnEmptyResponseWith(expectedStatusCode)
	client := NewAuthenticatedGitHubClient(context.TODO(), "", WithGraphQLAPIURL(gitHubGraphQLAPIMock.URL))

	// when
	deletedEntryID, err := client.UpdateIPAllowListEntry(context.TODO(), "some-entry-id", someIpAllowListEntryParameters)

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

func updateEntryResponseWith(expectedEntry IPAllowListEntry) string {
	res := fmt.Sprintf(updateEntryResponseTemplate, expectedEntry.ID, expectedEntry.AllowListValue, expectedEntry.IsActive, expectedEntry.Name, expectedEntry.CreatedAt.Format(gitHubTimeFormat), expectedEntry.UpdatedAt.Format(gitHubTimeFormat))
	return res
}
