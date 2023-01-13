package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/sync/semaphore"
	"io"
	"net/http"
	"sync"
)

const (
	defaultAPIURL = "https://api.github.com/graphql"
)

type ErrorWithStatusCode struct {
	StatusCode int
	message    string
}

func (e ErrorWithStatusCode) Error() string {
	return e.message
}

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
	http                 *http.Client
	concurrencySemaphore *semaphore.Weighted
	url                  string
	headers              map[string]string

	cacheEntries                  bool
	organizationEntriesCache      map[string][]*IPAllowListEntry
	organizationEntriesCacheMutex *sync.Mutex
	enterpriseEntriesCache        map[string][]*IPAllowListEntry
	enterpriseEntriesCacheMutex   *sync.Mutex
}

type ClientOptions struct {
	concurrency   int64
	graphQLAPIURL string
	headers       map[string]string
	cacheEntries  bool
}

type ClientOption func(options *ClientOptions)

// NewAuthenticatedGitHubClient creates a new authenticated client (using Personal Access Token (classic)) with given ClientOptions
func NewAuthenticatedGitHubClient(ctx context.Context, token string, opts ...ClientOption) *Client {
	authToken := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauthClient := oauth2.NewClient(ctx, authToken)

	return NewGitHubClient(oauthClient, opts...)
}

// NewGitHubClient creates a new client with a given http.Client and ClientOptions.
// This client does not handle authorization nor authentication, instead relies on http.Client implementation.
func NewGitHubClient(httpClient *http.Client, opts ...ClientOption) *Client {
	options := &ClientOptions{
		concurrency:   int64(1),
		graphQLAPIURL: defaultAPIURL,
		cacheEntries:  true,
	}
	for _, opt := range opts {
		opt(options)
	}

	c := &Client{
		http:                 httpClient,
		concurrencySemaphore: semaphore.NewWeighted(options.concurrency),
		url:                  options.graphQLAPIURL,
		headers:              options.headers,
	}

	if options.cacheEntries {
		c.cacheEntries = true
		c.organizationEntriesCache = make(map[string][]*IPAllowListEntry, 8)
		c.organizationEntriesCacheMutex = &sync.Mutex{}
		c.enterpriseEntriesCache = make(map[string][]*IPAllowListEntry, 8)
		c.enterpriseEntriesCacheMutex = &sync.Mutex{}
	}
	return c
}

// WithConcurrency determines maximum number of concurrent requests to the GitHub GraphQL API. Used to control rate limiting.
// concurrency must be >= 1, otherwise the option is ignored.
func WithConcurrency(concurrency int64) ClientOption {
	return func(options *ClientOptions) {
		if concurrency >= 1 {
			options.concurrency = concurrency
		}
	}
}

// WithGraphQLAPIURL sets GitHub's base GraphQL API URL.
func WithGraphQLAPIURL(graphQLAPIURL string) ClientOption {
	return func(options *ClientOptions) {
		options.graphQLAPIURL = graphQLAPIURL
	}
}

// WithHeaders adds additional HTTP headers
func WithHeaders(headers map[string]string) ClientOption {
	return func(options *ClientOptions) {
		options.headers = headers
	}
}

// WithEntriesCaching enables an entries cache. It reduces number of calls to GitHub's GraphQL API reducing rate limiting pressure.
func WithEntriesCaching() ClientOption {
	return func(options *ClientOptions) {
		options.cacheEntries = true
	}
}

// WithoutEntriesCaching disables an entries cache.
// Without caching client will perform multiple calls to GitHub's GraphQL API for each entry listing function call like GetOrganizationIPAllowListEntries.
// Also, it might put pressure on yours GitHub rate limit.
func WithoutEntriesCaching() ClientOption {
	return func(options *ClientOptions) {
		options.cacheEntries = false
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
	req, err := c.createRequestWithBody(ctx, reqData)
	if err != nil {
		return nil, err
	}

	res, err := c.doRequestWithConcurrency(ctx, req)
	if err != nil {
		return nil, err
	}

	gqlRes, err := handleGraphQLResponse(res)
	if err != nil {
		return nil, err
	}

	err = handleErrors(gqlRes)
	if err != nil {
		return nil, err
	}

	resData, err := toResponseData[T](gqlRes)
	if err != nil {
		return nil, err
	}
	return resData, nil
}

func (c *Client) createRequestWithBody(ctx context.Context, reqData GraphQLRequest) (*http.Request, error) {
	b, err := json.Marshal(reqData)
	if err != nil {
		return nil, errors.Wrap(err, "request marshalling error")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(b))
	if err != nil {
		return nil, errors.Wrap(err, "request error")
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func (c *Client) doRequestWithConcurrency(ctx context.Context, req *http.Request) (*http.Response, error) {
	err := c.concurrencySemaphore.Acquire(ctx, int64(1))
	if err != nil {
		return nil, errors.Wrap(err, "cannot acquire semaphore")
	}
	defer c.concurrencySemaphore.Release(int64(1))
	res, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http request call error")
	}
	return res, nil
}

func handleGraphQLResponse(res *http.Response) (*GraphQLResponse, error) {
	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, ErrorWithStatusCode{
			StatusCode: res.StatusCode,
			message:    fmt.Sprintf("response body read error: %s", err.Error()),
		}
	}

	if res.StatusCode >= 300 {
		return nil, ErrorWithStatusCode{
			StatusCode: res.StatusCode,
			message:    fmt.Sprintf("GitHub API response: %s", string(resBytes)),
		}
	}

	gqlRes := new(GraphQLResponse)
	err = json.Unmarshal(resBytes, gqlRes)
	if err != nil {
		return nil, ErrorWithStatusCode{
			StatusCode: res.StatusCode,
			message:    fmt.Sprintf("response unmarshalling error: %s", err.Error()),
		}
	}

	return gqlRes, err
}

func handleErrors(gqlRes *GraphQLResponse) error {
	if len(gqlRes.Errors) > 0 {
		var errs error
		for _, e := range gqlRes.Errors {
			errs = multierror.Append(errs, errors.New(e.Message))
		}
		return errs
	}
	return nil
}

func toResponseData[T any](gqlRes *GraphQLResponse) (*T, error) {
	resData := new(T)
	err := json.Unmarshal(gqlRes.Data, resData)
	if err != nil {
		return nil, errors.Wrap(err, "response unmarshalling error")
	}
	return resData, nil
}
