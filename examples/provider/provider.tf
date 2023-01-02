provider "githubipallowlist" {
  token        = "foo"
  organization = "your-org-name"
  base_url     = "https://your-github-enterprise-instance.com/graphql"
  concurrency  = 2
}
