package github

import (
	"context"
	"github.com/pkg/errors"
)

const getOrganizationIdQuery = `
query GetOrganizationId($organizationName: String!) {
  organization(login: $organizationName) {
    id
  }
}`

type GetOrganizationIdQueryResponse struct {
	Organization struct {
		Id string `json:"id"`
	} `json:"organization"`
}

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
}`

type GetOrganizationIPAllowListQueryResponse struct {
	Organization struct {
		IPAllowListEntries struct {
			Nodes    []*IPAllowListEntry `json:"nodes"`
			PageInfo PageInfo            `json:"pageInfo"`
		} `json:"ipAllowListEntries"`
	} `json:"organization"`
}

// GetOrganizationIPAllowListEntries retrieves IP allow list entries for a given organizationName.
// Method fetches all entries which might be a subject to rate limiting for allow lists with a big number of entries.
// Returns a slice of pointers to an entry as the API returns nil for entries managed on an enterprise level.
func (c *Client) GetOrganizationIPAllowListEntries(ctx context.Context, organizationName string) ([]*IPAllowListEntry, error) {
	var entries []*IPAllowListEntry
	var err error
	if c.cacheEntries {
		entries, err = c.getOrganizationIPAllowListEntriesWithCache(ctx, organizationName)
	} else {
		entries, err = c.getOrganizationIPAllowListEntries(ctx, organizationName)
	}

	return entries, err
}

func (c *Client) getOrganizationIPAllowListEntriesWithCache(ctx context.Context, organizationName string) ([]*IPAllowListEntry, error) {
	c.organizationEntriesCacheMutex.Lock()
	defer c.organizationEntriesCacheMutex.Unlock()

	var entries []*IPAllowListEntry
	var ok bool
	entries, ok = c.organizationEntriesCache[organizationName]
	if !ok {
		var err error
		entries, err = c.getOrganizationIPAllowListEntries(ctx, organizationName)
		if err != nil {
			return entries, errors.Wrap(err, "getOrganizationIPAllowListEntriesWithCache error")
		}

		c.organizationEntriesCache[organizationName] = entries
	}
	return entries, nil
}

func (c *Client) getOrganizationIPAllowListEntries(ctx context.Context, organizationName string) ([]*IPAllowListEntry, error) {
	reqData := GraphQLRequest{Query: getOrganizationIPAllowListEntriesQuery, Variables: map[string]any{"org": organizationName}}
	entries, err := paginate[GetOrganizationIPAllowListQueryResponse, IPAllowListEntry](ctx, c, reqData,
		func(t *GetOrganizationIPAllowListQueryResponse) []*IPAllowListEntry {
			return t.Organization.IPAllowListEntries.Nodes
		}, func(t *GetOrganizationIPAllowListQueryResponse) PageInfo {
			return t.Organization.IPAllowListEntries.PageInfo
		})

	if err != nil {
		return []*IPAllowListEntry{}, errors.Wrap(err, "getOrganizationIPAllowListEntries error")
	}
	return entries, nil
}

// GetOrganizationID fetches GitHub GraphQL API node_id for given organizationName.
func (c *Client) GetOrganizationID(ctx context.Context, organizationName string) (string, error) {
	reqData := GraphQLRequest{
		Query: getOrganizationIdQuery,
		Variables: map[string]any{
			"organizationName": organizationName,
		}}

	resData, err := doRequest[GetOrganizationIdQueryResponse](ctx, c, reqData)
	if err != nil {
		return "", errors.Wrap(err, "GetOrganizationID error")
	}

	return resData.Organization.Id, nil
}
