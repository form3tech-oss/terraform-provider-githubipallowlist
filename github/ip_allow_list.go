package github

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

type IPAllowListEntry struct {
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	AllowListValue string    `json:"allowListValue"`
	IsActive       bool      `json:"isActive"`
	Name           string    `json:"name"`
}

const createIPAllowListEntryMutation = `
mutation CreateIpAllowListEntry($ownerId: ID!, $name: String = "", $value: String!, $isActive: Boolean!) {
  createIpAllowListEntry(
    input: {ownerId: $ownerId, allowListValue: $value, isActive: $isActive, name: $name}
  ) {
    ipAllowListEntry {
      id
      createdAt
      updatedAt
      allowListValue
      isActive
      name
    }
  }
}`

const deleteIPAllowListEntryMutation = `
mutation DeleteIpAllowListEntry($entryId: ID!) {
  deleteIpAllowListEntry(input: {ipAllowListEntryId: $entryId}) {
    ipAllowListEntry {
      id
    }
  }
}
`

type DeleteUpAllowListEntryMutationResponse struct {
	DeleteIpAllowListEntry struct {
		IpAllowListEntry struct {
			ID string `json:"id"`
		} `json:"ipAllowListEntry"`
	} `json:"deleteIpAllowListEntry"`
}

type CreateIPAllowListEntryMutationResponse struct {
	CreateIPAllowListEntry struct {
		IPAllowListEntry IPAllowListEntry `json:"ipAllowListEntry"`
	} `json:"createIpAllowListEntry"`
}

// CreateIPAllowListEntry uses createIpAllowListEntry GraphQL mutation to create a new IP allow list entry for a given ownerID (organization or enterprise).
// Returns the newly created entry.
func (c *Client) CreateIPAllowListEntry(ctx context.Context, ownerID string, name string, value string, isActive bool) (*IPAllowListEntry, error) {
	reqData := GraphQLRequest{
		Query: createIPAllowListEntryMutation,
		Variables: map[string]any{
			"ownerId":  ownerID,
			"value":    value,
			"isActive": isActive,
		}}

	if name != "" {
		reqData.Variables["name"] = name
	}

	resData, err := doRequest[CreateIPAllowListEntryMutationResponse](ctx, c, reqData)
	if err != nil {
		return nil, errors.Wrap(err, "CreateIPAllowListEntry error")
	}

	return &resData.CreateIPAllowListEntry.IPAllowListEntry, nil
}

// DeleteIPAllowListEntry uses deleteIpAllowListEntry GraphQL mutation to delete an IP allow list entry with a given entryID.
// Returns entryID of the deleted entry.
func (c *Client) DeleteIPAllowListEntry(ctx context.Context, entryID string) (string, error) {
	reqData := GraphQLRequest{
		Query: deleteIPAllowListEntryMutation,
		Variables: map[string]any{
			"entryId": entryID,
		}}

	resData, err := doRequest[DeleteUpAllowListEntryMutationResponse](ctx, c, reqData)
	if err != nil {
		return "", errors.Wrap(err, "DeleteIPAllowListEntry error")
	}

	return resData.DeleteIpAllowListEntry.IpAllowListEntry.ID, nil
}
