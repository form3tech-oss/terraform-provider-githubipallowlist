package github

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const getOrganizationIDResponseTemplate = `{
    "data": {
        "organization": {
            "id": "%s"
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

func getOrganizationIDResponseWith(expectedOrganizationID string) string {
	return fmt.Sprintf(getOrganizationIDResponseTemplate, expectedOrganizationID)
}