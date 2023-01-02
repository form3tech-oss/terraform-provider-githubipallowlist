# terraform-provider-githubipallowlist

A Terraform provider for managing GitHub's IP allow list.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.3.*
- [Go](https://golang.org/doc/install) >= 1.19

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```sh
$ go install
```

This enables verifying your locally built provider using examples available in the `examples/` directory.
Note that you will first need to configure your shell to map our provider to the local build:

```sh
export TF_CLI_CONFIG_FILE=path/to/project/examples/dev.tfrc
```

An example file is available in our `examples` directory and resembles:

```hcl
provider_installation {
  dev_overrides {
    "from3tech-oss/githubipallowlist" = "~/go/bin/"
  }
  direct {}
}
```

See https://www.terraform.io/docs/cli/config/config-file.html for more details.

When running examples, you should spot the following warning to confirm you are using a local build:

```console
Warning: Provider development overrides are in effect

The following provider development overrides are set in the CLI configuration:
 - from3tech-oss/githubipallowlist in /Users/somegithuuser/go/bin
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/form3tech-oss/terraform-provider-githubipallowlist` to your Terraform provider:

```
go get github.com/form3tech-oss/terraform-provider-githubipallowlist
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

TBD

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (
see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin`
directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
