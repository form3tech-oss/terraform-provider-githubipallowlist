package github

import (
	"context"
	"github.com/pkg/errors"
)

const getOrganizationIPAllowListEntriesQuery = `
query GetOrganizationIpAllowListEntries($org: String!, $after: String) {
  organization(login: $org) {
    ipAllowListEntries(first: 100, after: $after) {
      nodes {
        allowListValue
        isActive
        name
        id
        createdAt
        updatedAt
      }
      pageInfo {
        hasNextPage
        startCursor
        endCursor
      }
    }
  }
}
`

type GetOrganizationIpAllowListQueryResponse struct {
	Organization struct {
		IpAllowListEntries struct {
			Nodes    []*IpAllowListEntry `json:"nodes"`
			PageInfo PageInfo            `json:"pageInfo"`
		} `json:"ipAllowListEntries"`
	} `json:"organization"`
}

func (c *Client) GetOrganizationIPAllowListEntries(ctx context.Context, organizationName string) ([]*IpAllowListEntry, error) {
	var entries []*IpAllowListEntry
	reqData := GraphQLRequest{Query: getOrganizationIPAllowListEntriesQuery, Variables: map[string]any{"org": organizationName}}
	entries, err := paginate[GetOrganizationIpAllowListQueryResponse, IpAllowListEntry](ctx, c, reqData,
		func(t *GetOrganizationIpAllowListQueryResponse) []*IpAllowListEntry {
			return t.Organization.IpAllowListEntries.Nodes
		}, func(t *GetOrganizationIpAllowListQueryResponse) PageInfo {
			return t.Organization.IpAllowListEntries.PageInfo
		})

	if err != nil {
		return []*IpAllowListEntry{}, errors.Wrap(err, "GetOrganizationIPAllowListEntries error")
	}

	return entries, nil
}
