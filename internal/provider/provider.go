package provider

import (
	"context"
	"github.com/form3tech-oss/terraform-provider-githubipallowlist/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("GITHUB_TOKEN", nil),
					Description: "Personal Access Token (classic). Defaults to a value of a GITHUB_TOKEN environmental variable.",
				},
				"organization": {
					Type:         schema.TypeString,
					Optional:     true,
					DefaultFunc:  schema.EnvDefaultFunc("GITHUB_ORGANIZATION", nil),
					Description:  "The GitHub organization name to manage. Defaults to a value of a GITHUB_ORGANIZATION environmental variable.",
					ExactlyOneOf: []string{"organization", "enterprise"},
				},
				"enterprise": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("GITHUB_ENTERPRISE", nil),
					Description: "The GitHub enterprise name to manage. Defaults to a value of a GITHUB_ENTERPRISE environmental variable.",
				},
				"base_url": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("GITHUB_BASE_URL", "https://api.github.com/graphql"),
					Description: "The GitHub base GraphQL API URL. Defaults to a value of a GITHUB_BASE_URL environmental variable.",
				},
				"concurrency": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     1,
					Description: "Concurrency of the client. Determines maximum number of concurrent requests to the GitHub GraphQL API. Used to control rate limiting. Default: 1.",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"githubipallowlist_ip_allow_list_entry": resourceGitHubIPAllowListEntry(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	github         *github.Client
	ownerName      string
	ownerID        string
	getEntriesFunc func(context.Context, string) ([]*github.IPAllowListEntry, error)
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		token := d.Get("token").(string)
		baseURL := d.Get("base_url").(string)
		concurrency := d.Get("concurrency").(int)
		organization := d.Get("organization").(string)
		enterprise := d.Get("enterprise").(string)

		userAgent := p.UserAgent("terraform-provider-githubipallowlist", version)

		ghc := github.NewAuthenticatedGitHubClient(ctx, token,
			github.WithGraphQLAPIURL(baseURL),
			github.WithConcurrency(int64(concurrency)),
			github.WithHeaders(map[string]string{"User-Agent": userAgent}),
		)

		var ownerID string
		var ownerName string
		var getEntriesFunc func(context.Context, string) ([]*github.IPAllowListEntry, error)
		if organization != "" {
			id, err := ghc.GetOrganizationID(ctx, organization)

			if err != nil {
				return nil, diag.FromErr(err)
			}

			ownerID = id
			ownerName = organization
			getEntriesFunc = ghc.GetOrganizationIPAllowListEntries

		}
		if enterprise != "" {
			id, err := ghc.GetEnterpriseID(ctx, enterprise)

			if err != nil {
				return nil, diag.FromErr(err)
			}

			ownerID = id
			ownerName = enterprise
			getEntriesFunc = ghc.GetEnterpriseIPAllowListEntries
		}

		return &apiClient{
			github:         ghc,
			ownerName:      ownerName,
			ownerID:        ownerID,
			getEntriesFunc: getEntriesFunc,
		}, nil
	}
}
