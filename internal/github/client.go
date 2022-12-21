package github

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/sync/semaphore"
	"io"
	"net/http"

	"github.com/hashicorp/go-multierror"
)

const (
	defaultApiURL = "https://api.github.com/graphql"
)

type Variables map[string]any

type GraphQLRequest struct {
	Query     string    `json:"query"`
	Variables Variables `json:"variables"`
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type Error struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Path []string `json:"path"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []Error         `json:"errors"`
}

type Client struct {
	http      *http.Client
	semaphore *semaphore.Weighted
	url       string
}

type ClientOptions struct {
	concurrency int64
}

type ClientOption func(options *ClientOptions)

func NewGitHubClient(ctx context.Context, token string, opts ...ClientOption) *Client {
	authToken := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	options := &ClientOptions{
		concurrency: int64(1),
	}
	for _, opt := range opts {
		opt(options)
	}

	oauthClient := oauth2.NewClient(ctx, authToken)

	return &Client{
		http:      oauthClient,
		semaphore: semaphore.NewWeighted(options.concurrency),
		url:       defaultApiURL,
	}
}

func WithConcurrency(concurrency int64) ClientOption {
	return func(options *ClientOptions) {
		options.concurrency = concurrency
	}
}

func paginate[T any, L any](ctx context.Context, c *Client, reqData GraphQLRequest, pageExtractor func(*T) []*L, pageInfoExtractor func(*T) PageInfo) ([]*L, error) {
	entries := make([]*L, 0, 10)
	hasNextPage := true
	endCursor := ""
	for hasNextPage {
		if endCursor != "" {
			reqData.Variables["after"] = endCursor
		}

		resData, err := doRequest[T](ctx, c, reqData)
		if err != nil {
			return entries, errors.Wrap(err, "pagination error")
		}

		entries = append(entries, pageExtractor(resData)...)

		pageInfo := pageInfoExtractor(resData)
		if pageInfo.HasNextPage {
			endCursor = pageInfo.EndCursor
		} else {
			hasNextPage = false
		}
	}

	return entries, nil
}

func doRequest[T any](ctx context.Context, c *Client, reqData GraphQLRequest) (*T, error) {
	b, err := json.Marshal(reqData)
	if err != nil {
		return nil, errors.Wrap(err, "request marshalling error")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(b))
	if err != nil {
		return nil, errors.Wrap(err, "request error")
	}
	err = c.semaphore.Acquire(ctx, int64(1))
	if err != nil {
		return nil, errors.Wrap(err, "cannot acquire semaphore")
	}
	defer c.semaphore.Release(int64(1))
	res, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "call error")
	}
	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "response body read error")
	}

	if res.StatusCode >= 300 {
		return nil, errors.Wrapf(err, "GitHub API response: %d - %s", res.StatusCode, string(resBytes))
	}

	gqlRes := new(GraphQLResponse)
	err = json.Unmarshal(resBytes, gqlRes)
	if err != nil {
		return nil, errors.Wrap(err, "response unmarshalling error")
	}

	if len(gqlRes.Errors) > 0 {
		var errs error
		for _, e := range gqlRes.Errors {
			errs = multierror.Append(errs, errors.New(e.Message))
		}
		return nil, errs
	}

	resData := new(T)
	err = json.Unmarshal(gqlRes.Data, resData)
	if err != nil {
		return nil, errors.Wrap(err, "response unmarshalling error")
	}
	return resData, nil
}
