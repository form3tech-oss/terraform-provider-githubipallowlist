package main

import (
	"flag"
	"github.com/form3tech-oss/terraform-provider-githubipallowlist/internal/provider"
	"runtime/debug"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	buildInfo, ok := debug.ReadBuildInfo()
	version := "dev"
	if ok {
		version = buildInfo.Main.Version
	}

	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        debugMode,
		ProviderFunc: provider.New(version),
	}

	plugin.Serve(opts)
}
