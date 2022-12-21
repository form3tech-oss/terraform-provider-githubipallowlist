package github

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

type IpAllowListEntry struct {
	Id             string    `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	AllowListValue string    `json:"allowListValue"`
	IsActive       bool      `json:"isActive"`
	Name           string    `json:"name"`
}

const createIpAllowListEntryMutation = `
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

type CreateIpAllowListEntryMutationResponse struct {
	CreateIpAllowListEntry struct {
		IpAllowListEntry IpAllowListEntry `json:"ipAllowListEntry"`
	} `json:"createIpAllowListEntry"`
}

const updateIpAllowListEntryMutation = `
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

type UpdateIpAllowListEntryMutationResponse struct {
	UpdateIpAllowListEntry struct {
		IpAllowListEntry IpAllowListEntry `json:"ipAllowListEntry"`
	} `json:"updateIpAllowListEntry"`
}

const deleteIpAllowListEntryMutation = `
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
			Id string `json:"id"`
		} `json:"ipAllowListEntry"`
	} `json:"deleteIpAllowListEntry"`
}

func (c *Client) CreateIpAllowListEntry(ctx context.Context, ownerID string, name string, value string, isActive bool) (*IpAllowListEntry, error) {
	reqData := GraphQLRequest{
		Query: createIpAllowListEntryMutation,
		Variables: map[string]any{
			"ownerId":  ownerID,
			"value":    value,
			"isActive": isActive,
		}}

	if name != "" {
		reqData.Variables["name"] = name
	}

	resData, err := doRequest[CreateIpAllowListEntryMutationResponse](ctx, c, reqData)
	if err != nil {
		return nil, errors.Wrap(err, "CreateIpAllowListEntry error")
	}

	return &resData.CreateIpAllowListEntry.IpAllowListEntry, nil
}

func (c *Client) UpdateIpAllowListEntry(ctx context.Context, entryID string, name string, value string, isActive bool) (*IpAllowListEntry, error) {
	reqData := GraphQLRequest{
		Query: updateIpAllowListEntryMutation,
		Variables: map[string]any{
			"entryId":  entryID,
			"name":     name,
			"value":    value,
			"isActive": isActive,
		}}

	resData, err := doRequest[UpdateIpAllowListEntryMutationResponse](ctx, c, reqData)
	if err != nil {
		return nil, errors.Wrap(err, "UpdateIpAllowListEntry error")
	}

	return &resData.UpdateIpAllowListEntry.IpAllowListEntry, nil
}

func (c *Client) DeleteIpAllowListEntry(ctx context.Context, entryID string) (string, error) {
	reqData := GraphQLRequest{
		Query: deleteIpAllowListEntryMutation,
		Variables: map[string]any{
			"entryId": entryID,
		}}

	resData, err := doRequest[DeleteUpAllowListEntryMutationResponse](ctx, c, reqData)
	if err != nil {
		return "", errors.Wrap(err, "DeleteIpAllowListEntry error")
	}

	return resData.DeleteIpAllowListEntry.IpAllowListEntry.Id, nil
}
