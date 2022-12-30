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

type CreateIPAllowListEntryMutationResponse struct {
	CreateIPAllowListEntry struct {
		IPAllowListEntry IPAllowListEntry `json:"ipAllowListEntry"`
	} `json:"createIpAllowListEntry"`
}

const updateIPAllowListEntryMutation = `
mutation UpdateIpAllowListEntry($entryId: ID!, $name: String!, $value: String!, $isActive: Boolean!) {
  updateIpAllowListEntry(
    input: {ipAllowListEntryId: $entryId, allowListValue: $value, isActive: $isActive, name: $name}
  ) {
      ipAllowListEntry {
      allowListValue
      createdAt
      id
      isActive
      name
      updatedAt
    }
  }
}`

type UpdateIPAllowListEntryMutationResponse struct {
	UpdateIPAllowListEntry struct {
		IPAllowListEntry IPAllowListEntry `json:"ipAllowListEntry"`
	} `json:"updateIpAllowListEntry"`
}

const deleteIPAllowListEntryMutation = `
mutation DeleteIpAllowListEntry($entryId: ID!) {
  deleteIpAllowListEntry(input: {ipAllowListEntryId: $entryId}) {
    ipAllowListEntry {
      id
    }
  }
}`

type DeleteIPAllowListEntryMutationResponse struct {
	DeleteIPAllowListEntry struct {
		IPAllowListEntry struct {
			Id string `json:"id"`
		} `json:"ipAllowListEntry"`
	} `json:"deleteIpAllowListEntry"`
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

// UpdateIPAllowListEntry uses updateIpAllowListEntry GraphQL mutation to update IP allow list entry. It replaces all entry parameters with ones provided in arguments.
// Returns the updated entry.
func (c *Client) UpdateIPAllowListEntry(ctx context.Context, entryID string, name string, value string, isActive bool) (*IPAllowListEntry, error) {
	reqData := GraphQLRequest{
		Query: updateIPAllowListEntryMutation,
		Variables: map[string]any{
			"entryId":  entryID,
			"name":     name,
			"value":    value,
			"isActive": isActive,
		}}

	resData, err := doRequest[UpdateIPAllowListEntryMutationResponse](ctx, c, reqData)
	if err != nil {
		return nil, errors.Wrap(err, "UpdateIPAllowListEntry error")
	}

	return &resData.UpdateIPAllowListEntry.IPAllowListEntry, nil
}

// DeleteIPAllowListEntry uses deleteIPAllowListEntryMutation GraphQL mutation to delete IP allow list entry with a given entryID.
// Returns entryID of the deleted entry.
func (c *Client) DeleteIPAllowListEntry(ctx context.Context, entryID string) (string, error) {
	reqData := GraphQLRequest{
		Query: deleteIPAllowListEntryMutation,
		Variables: map[string]any{
			"entryId": entryID,
		}}

	resData, err := doRequest[DeleteIPAllowListEntryMutationResponse](ctx, c, reqData)
	if err != nil {
		return "", errors.Wrap(err, "DeleteIPAllowListEntry error")
	}

	return resData.DeleteIPAllowListEntry.IPAllowListEntry.Id, nil
}
