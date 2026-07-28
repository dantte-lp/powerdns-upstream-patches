// Command terraform-provider-powerdns serves the PowerDNS Terraform provider.
//
// The provider is served through a multiplexer while the migration from
// terraform-plugin-sdk/v2 to terraform-plugin-framework is in progress
// (docs/adr/0003-plugin-framework-migration.md). Both halves speak protocol
// 6.0: the SDKv2 provider is lifted from 5.0 by tf5to6server, and the
// framework provider speaks 6.0 natively.
//
// Resources move one at a time. A resource must exist in exactly one half —
// the mux server refuses to start if a type name is served by both — so each
// port is a single commit, and the provider stays shippable throughout.
//
// When the last resource has moved, this file collapses to a plain
// providerserver.Serve and terraform-plugin-mux is dropped (plan S4-10).
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/dantte-lp/terraform-provider-powerdns/internal/provider"
	"github.com/dantte-lp/terraform-provider-powerdns/powerdns"
)

// registryAddress is the provider's address in the Terraform Registry. It must
// match the source in a consumer's required_providers block.
const registryAddress = "registry.terraform.io/dantte-lp/powerdns"

// version is set by the linker at release time; see the goreleaser ldflags.
var version = "dev"

func main() {
	ctx := context.Background()

	var debug bool
	flag.BoolVar(&debug, "debug", false,
		"run the provider in debug mode, for use with a debugger such as delve")
	flag.Parse()

	// Lift the SDKv2 provider from protocol 5.0 to 6.0 so it can be muxed with
	// the framework provider.
	upgraded, err := tf5to6server.UpgradeServer(ctx, powerdns.Provider().GRPCProvider)
	if err != nil {
		log.Fatalf("upgrading the SDKv2 provider to protocol 6: %v", err)
	}

	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(provider.New(version)()),
		func() tfprotov6.ProviderServer { return upgraded },
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatalf("creating the mux server: %v", err)
	}

	var opts []tf6server.ServeOpt
	if debug {
		opts = append(opts, tf6server.WithManagedDebug())
	}

	if err := tf6server.Serve(registryAddress, muxServer.ProviderServer, opts...); err != nil {
		log.Fatalf("serving the provider: %v", err)
	}
}
