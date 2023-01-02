package github

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

const (
	timeFormat = "2006-01-02T15:04:05Z"
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

func createEntryResponseWith(expectedEntry IPAllowListEntry) string {
	res := fmt.Sprintf(createEntryResponseTemplate, expectedEntry.ID, expectedEntry.CreatedAt.Format(timeFormat), expectedEntry.UpdatedAt.Format(timeFormat), expectedEntry.AllowListValue, expectedEntry.IsActive, expectedEntry.Name)
	return res
}
