package github

import (
	"context"
	"github.com/pkg/errors"
)

const getEnterpriseIDQuery = `
query GetEnterpriseId($enterpriseName: String!) {
  enterprise(slug: $enterpriseName) {
    id
  }
}`

type GetEnterpriseIDQueryResponse struct {
	Enterprise struct {
		ID string `json:"id"`
	} `json:"enterprise"`
}

const getEnterpriseIPAllowListEntriesQuery = `
query GetEnterpriseId($enterpriseName: String!, $after: String) {
  enterprise(slug: $enterpriseName) {
    ownerInfo {
      ipAllowListEntries(first: 100, after: $after) {
        nodes {
          id
          allowListValue
          name
          isActive
          createdAt
          updatedAt
        }
        pageInfo {
          endCursor
          hasNextPage
          startCursor
        }
      }
    }
  }
}`

type GetEnterpriseIPAllowListQueryResponse struct {
	Enterprise struct {
		OwnerInfo struct {
			IPAllowListEntries struct {
				Nodes    []*IPAllowListEntry `json:"nodes"`
				PageInfo PageInfo            `json:"pageInfo"`
			} `json:"ipAllowListEntries"`
		} `json:"ownerInfo"`
	} `json:"enterprise"`
}

// GetEnterpriseIPAllowListEntries retrieves IP allow list entries for a given enterpriseName.
func (c *Client) GetEnterpriseIPAllowListEntries(ctx context.Context, enterpriseName string) ([]*IPAllowListEntry, error) {
	var entries []*IPAllowListEntry
	var err error
	if c.cacheEntries {
		entries, err = c.getEnterpriseIPAllowListEntriesWithCache(ctx, enterpriseName)
	} else {
		entries, err = c.getEnterpriseIPAllowListEntries(ctx, enterpriseName)
	}

	return entries, err
}

func (c *Client) getEnterpriseIPAllowListEntriesWithCache(ctx context.Context, enterpriseName string) ([]*IPAllowListEntry, error) {
	c.enterpriseEntriesCacheMutex.Lock()
	defer c.enterpriseEntriesCacheMutex.Unlock()

	var entries []*IPAllowListEntry
	var ok bool
	entries, ok = c.enterpriseEntriesCache[enterpriseName]
	if !ok {
		var err error
		entries, err = c.getEnterpriseIPAllowListEntries(ctx, enterpriseName)
		if err != nil {
			return entries, errors.Wrap(err, "getEnterpriseIPAllowListEntriesWithCache error")
		}

		c.enterpriseEntriesCache[enterpriseName] = entries
	}
	return entries, nil
}

func (c *Client) getEnterpriseIPAllowListEntries(ctx context.Context, enterpriseName string) ([]*IPAllowListEntry, error) {
	reqData := GraphQLRequest{Query: getEnterpriseIPAllowListEntriesQuery, Variables: map[string]any{"org": enterpriseName}}
	entries, err := paginate[GetEnterpriseIPAllowListQueryResponse, IPAllowListEntry](ctx, c, reqData,
		func(t *GetEnterpriseIPAllowListQueryResponse) []*IPAllowListEntry {
			return t.Enterprise.OwnerInfo.IPAllowListEntries.Nodes
		}, func(t *GetEnterpriseIPAllowListQueryResponse) PageInfo {
			return t.Enterprise.OwnerInfo.IPAllowListEntries.PageInfo
		})

	if err != nil {
		return []*IPAllowListEntry{}, errors.Wrap(err, "getEnterpriseIPAllowListEntries error")
	}
	return entries, nil
}

// GetEnterpriseID fetches GitHub GraphQL API node_id for given enterpriseName.
func (c *Client) GetEnterpriseID(ctx context.Context, enterpriseName string) (string, error) {
	reqData := GraphQLRequest{
		Query: getEnterpriseIDQuery,
		Variables: map[string]any{
			"enterpriseName": enterpriseName,
		}}

	resData, err := doRequest[GetEnterpriseIDQueryResponse](ctx, c, reqData)
	if err != nil {
		return "", errors.Wrap(err, "GetEnterpriseID error")
	}

	return resData.Enterprise.ID, nil
}
